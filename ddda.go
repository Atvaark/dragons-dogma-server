package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/atvaark/dragons-dogma-server/cmd"
)

const AppVersion = "0.0.1"

func main() {
	app := cli.NewApp()
	app.Name = "Dragon's Dogma Server"
	app.Version = AppVersion
	app.Commands = []cli.Command{
		cmd.WebCommand,
	}

	app.Run(os.Args)
}
