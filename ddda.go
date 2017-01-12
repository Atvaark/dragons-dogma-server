package main

import (
	"os"

	"github.com/atvaark/dragons-dogma-server/cmd"
)

func main() {
	app := cmd.App{}
	app.Name = "Dragon's Dogma Server"
	app.Commands = []cmd.Command{
		cmd.WebCommand,
		cmd.CertCommand,
	}

	app.Run(os.Args)
}
