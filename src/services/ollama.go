package services

import (
	"encoding/json"
	"time"

	"github.com/LanceLRQ/ollama-watchdog/configs"
	"github.com/gofiber/fiber/v2"
)

func OllamaPSWatcher(cfg *configs.ServerConfigStruct, callback func(fiber.Map)) {
	ticker := time.NewTicker(time.Duration(time.Second))
	defer ticker.Stop()

	for range ticker.C {
		callback(GetOllamaPS(cfg))
	}
}

func ollamaAPIAgent(host string) (fiber.Map, error) {
	agent := fiber.Get(host + "/api/ps")
	agent.Timeout(time.Second)
	_, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	var resp fiber.Map
	err := json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Get Ollama Process Status
func GetOllamaPS(cfg *configs.ServerConfigStruct) fiber.Map {
	if len(cfg.OllamaListens) > 0 {
		responses := make([]fiber.Map, len(cfg.OllamaListens))
		for i, host := range cfg.OllamaListens {
			resp, err := ollamaAPIAgent(host)
			serviceName := ""
			if len(cfg.OllamaServices) > i {
				serviceName = cfg.OllamaServices[i]
			}
			if err != nil {
				responses[i] = fiber.Map{
					"server":       host,
					"service_name": serviceName,
					"status":       false,
					"data":         nil,
				}
				continue
			}
			responses[i] = fiber.Map{
				"server":       host,
				"service_name": serviceName,
				"status":       true,
				"data":         resp,
			}
		}
		return fiber.Map{
			"status":  true,
			"data":    responses,
			"message": "",
		}
	}
	return fiber.Map{
		"status":  false,
		"data":    nil,
		"message": "没有配置ollama监听地址",
	}

}
