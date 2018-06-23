package main

import (
	"os"
	"path/filepath"

	"gopkg.in/urfave/cli.v2"
)

const (
	appName = "iijmio-checker"
)

func main() {
	new().Run(os.Args)
}

func new() *cli.App {
	return &cli.App{
		Name:  appName,
		Usage: "Checker for usage of IIJmio SIM",
		Commands: []*cli.Command{
			{
				Name:    "auth",
				Aliases: []string{"a"},
				Usage:   "Launch the server to auth IIJmio API",
				Action:  auth,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "Filename for config",
				DefaultText: filepath.Join(
					os.Getenv("HOME"), ".config", appName, "config.json",
				),
			},
			&cli.StringFlag{
				Name:  "session-config",
				Usage: "Filename for config of sessions in auth server",
				DefaultText: filepath.Join(
					os.Getenv("HOME"), ".config", appName, "session-config.gob",
				),
			},
		},
	}
}