package cmd

import "fmt"

var CertCommand = Command{
	Name:        "cert",
	Description: "Creates a self signed certificate for the specified host",
	Flags: []Flag{
		{Name: "host", Value: "localhost"},
	},
	action: runCert,
	args: func() commandArgs {
		return &webArgs{}
	},
}

type certArgs struct {
	host string
}

func (a *certArgs) init(flags []Flag, args []string) error {
	// TODO: Read args
	a.host = "localhost"
	return nil
}

func runCert(ctx *context) {
	fmt.Println("cert")
	args := ctx.args.(*certArgs)
	fmt.Println(args)
}
