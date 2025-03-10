package services

import (
	"encoding/json"
	"time"

	"github.com/LanceLRQ/ollama-watchdog/configs"
	"github.com/gofiber/fiber/v2"
)

func OllamaPSWatcher(cfg *configs.ServerConfigStruct, callback func(fiber.Map)) {
	ticker := time.NewTicker(time.Duration(0.5 * float64(time.Second)))
	defer ticker.Stop()

	for range ticker.C {
		callback(GetOllamaPS(cfg))
	}
}

// Get Ollama Process Status
func GetOllamaPS(cfg *configs.ServerConfigStruct) fiber.Map {
	agent := fiber.Get(cfg.OllamaListen + "/api/ps")
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return fiber.Map{
			"status": false,
			"code":   statusCode,
			"error":  errs,
		}
	}

	var resp fiber.Map
	err := json.Unmarshal(body, &resp)
	if err != nil {
		return fiber.Map{
			"status": false,
			"code":   statusCode,
			"error":  errs,
		}
	}

	return fiber.Map{
		"status": true,
		"code":   statusCode,
		"data":   resp,
	}
}
