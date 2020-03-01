package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli"

	"github.com/raystlin/ftpsync/cmd"
	"github.com/raystlin/ftpsync/config"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "Activate debug log level",
			EnvVar: "FTP_SYNC_DEBUG",
		},
		cli.StringFlag{
			Name:   "config",
			Usage:  "JSON configuration file",
			EnvVar: "FTP_SYNC_CONFIG_FILE",
		},
	}

	app.Commands = []cli.Command{
		cmd.OnceCmd,
		cmd.DaemonCmd,
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		conf, err := config.ReadConfig(c.String("config"))
		if err != nil {
			log.WithFields(log.Fields{
				"error":       err,
				"config-file": c.String("config"),
			}).Error("Could not read the config file")
			return err
		}

		c.App.Metadata["config"] = conf

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Info("Done!!")
	}
}
