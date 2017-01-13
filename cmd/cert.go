package cmd

import (
	"flag"
	"fmt"
)

const (
	hostFlagName    = "host"
	hostFlagDefault = "localhost"
)

var CertCommand = Command{
	Name:        "cert",
	Description: "Creates a self signed certificate for the specified host",
	Flags: []Flag{
		{Name: hostFlagName, Value: hostFlagDefault},
	},
	action: runCert,
	args: func() commandArgs {
		return &certArgs{}
	},
}

type certArgs struct {
	host string
}

func (a *certArgs) init(flags []Flag, args []string) error {
	fls := flag.FlagSet{}
	fls.StringVar(&a.host, hostFlagName, hostFlagDefault, "")
	err := fls.Parse(args)
	if err != nil {
		return err
	}

	return nil
}

func runCert(ctx *context) {
	fmt.Println("cert")
	args := ctx.args.(*certArgs)
	fmt.Println(args)
}
