package cmd

import (
	"fmt"
	"strings"

	"github.com/LanceLRQ/ollama-watchdog/configs"
	"github.com/urfave/cli/v2"
)

func ConfigCommand() *cli.Command {
	return &cli.Command{
		Name:    "config",
		Aliases: []string{"cfg"},
		Usage:   "修改配置",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       configs.GetDefaultServerConfigPath(),
				Usage:       "配置文件路径",
				DefaultText: configs.GetDefaultServerConfigPath(),
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:  "get",
				Usage: "获取当前配置的值",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return fmt.Errorf("参数错误")
					}

					key := c.Args().Get(0)

					cfg, err := configs.ReadServerConfig(c.String("config"))
					if err != nil {
						return err
					}
					value, err := configs.GetConfigValue(cfg, key)
					if err != nil {
						return err
					}
					fmt.Printf("%v\n", value.Interface())
					return nil
				},
			},
			{
				Name:  "set",
				Usage: "设置配置值",
				Action: func(c *cli.Context) error {
					if c.NArg() < 2 {
						return fmt.Errorf("参数错误")
					}

					key := c.Args().Get(0)
					value := c.Args().Get(1)

					if key == "ollama_listen" {
						return fmt.Errorf("ollama_listen 已被弃用，请使用 ollama_listens")
					}

					cfg, err := configs.ReadServerConfig(c.String("config"))
					if err != nil {
						return err
					}

					fmt.Printf("修改配置文件：%s\n", c.String("config"))
					fmt.Printf("%s -> %s\n", key, value)

					err = configs.SetConfigValue(cfg, key, value)
					if err != nil {
						return err
					}

					// 简单校验一下格式
					if key == "ollama_listens" {
						for _, v := range cfg.OllamaListens {
							if !strings.HasPrefix(v, "http://") && !strings.HasPrefix(v, "https://") {
								return fmt.Errorf("地址 %s 必须以 http:// 或 https:// 开头", v)
							}
						}
					}

					// 写入
					err = configs.WriteServerConfig(c.String("config"), cfg)
					if err != nil {
						return err
					}
					fmt.Println("配置修改成功！")
					return nil
				},
			},
		},
	}
}
