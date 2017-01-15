package cmd

import (
	"fmt"

	"github.com/urfave/cli"
)

const (
	hostFlagName    = "host"
	hostFlagDefault = "localhost"
)

var CertCommand = cli.Command{
	Name:        "cert",
	Description: "Creates a self signed certificate for the specified host",
	Flags: []cli.Flag{
		cli.StringFlag{Name: hostFlagName, Value: hostFlagDefault},
	},
	Action: runCert,
}

type certConfig struct {
	host string
}

func runCert(ctx *cli.Context) {
	var cfg certConfig
	cfg.host = ctx.String(hostFlagName)
	fmt.Println("cert")
	fmt.Println(cfg)
}
