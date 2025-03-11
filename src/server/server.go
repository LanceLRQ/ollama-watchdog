package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LanceLRQ/ollama-watchdog/configs"
	"github.com/LanceLRQ/ollama-watchdog/models"
	"github.com/LanceLRQ/ollama-watchdog/services"
	"github.com/LanceLRQ/ollama-watchdog/utils"
	"github.com/dgraph-io/badger/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/websocket/v2"
)

func StartHttpServer(cfg *configs.ServerConfigStruct) error {
	var nvidiaResp models.NvidiaSMIResponse
	var ollamaPSResp fiber.Map

	GPUSampleDB, err := utils.OpenBadgerDB(cfg.GPUSampleDB)
	if err != nil {
		return fmt.Errorf("failed to open badger db: %w", err)
	}
	defer GPUSampleDB.Close()

	go services.NvidiaSMIWatcher(func(response models.NvidiaSMIResponse) {
		nvidiaResp = response
		services.SaveSampleToDB(GPUSampleDB, response)
	})
	go services.OllamaPSWatcher(cfg, func(response fiber.Map) {
		ollamaPSResp = response
	})

	app := fiber.New()
	// 使用 CORS 中间件
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",                              // 允许的域名
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH", // 允许的 HTTP 方法
		AllowHeaders: "Origin, Content-Type, Accept",   // 允许的请求头
	}))

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

	app.Get("/api/nvidia/history", func(c *fiber.Ctx) error {
		r := c.QueryInt("range", 120)

		start := time.Now().Add(time.Duration(-r) * time.Second).Unix()
		responstList := make([]models.NvidiaSMIResponse, 0)

		err = GPUSampleDB.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Seek(fmt.Appendf(nil, "gpu:%d", start)); it.Valid(); it.Next() {
				item := it.Item()
				v, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				var nvidiaRespLog models.NvidiaSMIResponse
				err = json.Unmarshal(v, &nvidiaRespLog)
				if err != nil {
					return err
				}
				responstList = append(responstList, nvidiaRespLog)
			}
			return nil
		})
		if err != nil {
			return c.JSON(fiber.Map{
				"status":  false,
				"message": err.Error(),
			})
		}
		return c.JSON(fiber.Map{
			"status": true,
			"data":   responstList,
		})
	})
	app.Get("/api/nvidia/now", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": true,
			"data":   nvidiaResp,
		})
	})

	app.Post("/api/kill", func(c *fiber.Ctx) error {
		data := new(struct {
			Type string `form: "type"`
			PID  int    `form: "pid"`
			Name string `form: "name"`
		})
		if err := c.BodyParser(&data); err != nil {
			return err
		}
		if data.Type == "ollama" {
			if data.Name != "" {
				err := utils.TerminateOllamaProcess(data.Name)
				if err != nil {
					return c.JSON(fiber.Map{"status": false, "message": "Failed to terminate process"})
				}
				return c.JSON(fiber.Map{"status": true})
			}
			return c.JSON(fiber.Map{"status": false, "message": "ollama model name is required"})
		} else if data.Type == "process" {
			if data.PID != 0 {

				err := utils.TerminateProcess(data.PID)
				if err != nil {
					return c.JSON(fiber.Map{"status": false, "message": "Failed to terminate process"})
				}
				return c.JSON(fiber.Map{"status": true})
			} else {
				return c.JSON(fiber.Map{"status": false, "message": "pid is required"})
			}
		}
		return c.JSON(fiber.Map{"status": false, "message": "type is not supported"})
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
