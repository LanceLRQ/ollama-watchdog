package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LanceLRQ/ollama-watchdog/models"
	"github.com/LanceLRQ/ollama-watchdog/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func startHttpServer() {
	app := fiber.New()

	var nvidiaResp models.NvidiaSMIResponse
	go services.NvidiaSMIWatcher(func(response models.NvidiaSMIResponse) {
		nvidiaResp = response
	})

	// WebSocket端点
	app.Get("/api/ws", websocket.New(func(c *websocket.Conn) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			jsonData, err := json.Marshal(nvidiaResp)
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

	// 静态文件服务
	app.Static("/", "../web/dist")

	app.Listen(":23333")
}
