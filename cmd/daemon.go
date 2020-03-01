package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/raystlin/ftpsync/config"
	"github.com/raystlin/ftpsync/sync"
)

const (
	IntervalFlag = "sync-interval"
)

var DaemonCmd = cli.Command{
	Name:  "once",
	Usage: "Execute a simple sync",
	Flags: []cli.Flag{
		cli.DurationFlag{
			Name:   IntervalFlag,
			Usage:  "Time to wait between syncs",
			Value:  6 * time.Hour,
			EnvVar: "FTP_SYNC_INTERVAL",
		},
	},
	Action: fxDaemon,
}

func fxDaemon(ctx *cli.Context) error {

	conf := config.FromContext(ctx)
	for {
		err := sync.Sync(conf)
		if err != nil {
			log.Error(err)
		}
		time.Sleep(ctx.Duration(IntervalFlag))
	}
}
