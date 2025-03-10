package cmd

import (
	"log"
	"os"

	"github.com/LanceLRQ/ollama-watchdog/configs"
	"github.com/LanceLRQ/ollama-watchdog/server"
	"github.com/urfave/cli/v2"
)

func CommandEntry() {
	app := &cli.App{
		Name:  "ollama-watchdog",
		Usage: "Ollama 监控小插件",
		Commands: []*cli.Command{
			ConfigCommand(),
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "server.yaml",
				Usage:   "配置文件路径",
			},
		},
		Action: func(c *cli.Context) error {
			cfg, err := configs.ReadServerConfig(c.String("config"))
			if err != nil {
				return err
			}
			return server.StartHttpServer(cfg)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
