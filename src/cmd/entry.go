package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/LanceLRQ/ollama-watchdog/configs"
	"github.com/LanceLRQ/ollama-watchdog/server"
	"github.com/urfave/cli/v2"
)

func CommandEntry(version string) {
	app := &cli.App{
		Name:  "ollama-watchdog",
		Usage: "Ollama 监控小插件",
		Commands: []*cli.Command{
			ConfigCommand(),
			&cli.Command{
				Name:  "serve",
				Usage: "启动监控服务",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "config",
						Aliases:     []string{"c"},
						Value:       configs.GetDefaultServerConfigPath(),
						Usage:       "配置文件路径",
						DefaultText: configs.GetDefaultServerConfigPath(),
					},
				},
				Action: func(c *cli.Context) error {
					cfg, err := configs.ReadServerConfig(c.String("config"))
					if err != nil {
						return err
					}
					return server.StartHttpServer(cfg)
				},
			},
			&cli.Command{
				Name:  "clean",
				Usage: "清理GPU监控数据",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "config",
						Aliases:     []string{"c"},
						Value:       configs.GetDefaultServerConfigPath(),
						Usage:       "配置文件路径",
						DefaultText: configs.GetDefaultServerConfigPath(),
					},
				},
				Action: func(c *cli.Context) error {
					// 清理逻辑
					cfg, err := configs.ReadServerConfig(c.String("config"))
					if err != nil {
						return err
					}
					fmt.Printf("即将清理目录：%s\n请先退出服务，然后再操作！\n是否继续清理[y/n]:", cfg.GPUSampleDB)

					var input string
					fmt.Scanf("%s", &input)
					if input == "y" || input == "Y" {
						if cfg.GPUSampleDB != "" {
							err = os.RemoveAll(cfg.GPUSampleDB)
							if err != nil {
								return fmt.Errorf("清理目录失败，请自行操作")
							}
							fmt.Println("清理完毕！")
						}
					}
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Printf("Ollama Watcher (built: %s)\n", version)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
