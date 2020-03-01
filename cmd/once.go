package cmd

import (
	"github.com/urfave/cli"

	"github.com/raystlin/ftpsync/config"
	"github.com/raystlin/ftpsync/sync"
)

var OnceCmd = cli.Command{
	Name:   "once",
	Usage:  "Execute a simple sync",
	Action: fxSync,
}

func fxSync(ctx *cli.Context) error {

	conf := config.FromContext(ctx)

	return sync.Sync(conf)
}
