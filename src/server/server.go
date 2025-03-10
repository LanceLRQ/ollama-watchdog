package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LanceLRQ/ollama-watchdog/configs"
	"github.com/LanceLRQ/ollama-watchdog/models"
	"github.com/LanceLRQ/ollama-watchdog/services"
	"github.com/LanceLRQ/ollama-watchdog/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/websocket/v2"
)

func StartHttpServer(cfg *configs.ServerConfigStruct) error {
	app := fiber.New()

	var nvidiaResp models.NvidiaSMIResponse
	var ollamaPSResp fiber.Map
	go services.NvidiaSMIWatcher(func(response models.NvidiaSMIResponse) {
		nvidiaResp = response
	})
	go services.OllamaPSWatcher(cfg, func(response fiber.Map) {
		ollamaPSResp = response
	})

	// WebSocket服务
	app.Get("/api/realtime", websocket.New(func(c *websocket.Conn) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			jsonData, err := json.Marshal(fiber.Map{
				"nvidia": nvidiaResp,
				"ollama": ollamaPSResp,
			})
			if err != nil {
				fmt.Println("JSON marshal error:", err)
				continue
			}

			if err := c.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				fmt.Println("Write error:", err)
				break
			}
		}
	}))

	app.Get("/api/nvidia", func(c *fiber.Ctx) error {
		return c.JSON(nvidiaResp)
	})
	
	app.Post("/api/kill", func(c *fiber.Ctx) error {
		var data fiber.Map
		if err := c.BodyParser(&data); err != nil {
			return err
		}
		if data["type"] != nil && data["type"].(string) == "ollama"  {
			if data["name"] != nil {
				if name, ok := data["name"].(string); ok {
					err := utils.TerminateOllamaProcess(name)
					if err != nil {	
						return c.JSON(fiber.Map{"status": false, "message": "Failed to terminate process"})
					}
					return c.JSON(fiber.Map{"status": true})
				}
			}
			return c.JSON(fiber.Map{"status": false, "message": "ollama model name is required"})
		} else {
			if data["pid"] != nil {
				pid := data["pid"].(int)
				
				err := utils.TerminateProcess(pid)
				if err != nil {	
					return c.JSON(fiber.Map{"status": false, "message": "Failed to terminate process"})
				}
				return c.JSON(fiber.Map{"status": true})
			} else {
				return c.JSON(fiber.Map{"status": false, "message": "pid is required"})
			}
		}
	})

	app.Get("/api/ollama/*", func(c *fiber.Ctx) error {
		ollamaApiPath := c.Params("*")
		if err := proxy.DoDeadline(c, cfg.OllamaListen+"/api/"+ollamaApiPath, time.Now().Add(time.Minute)); err != nil {
			return err
		}
		// Remove Server header from response
		c.Response().Header.Del(fiber.HeaderServer)
		return nil
	})

	// 静态文件服务
	// app.Static("/", "../web/dist")

	app.Listen(cfg.Listen)

	return nil
}
