package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/urfave/cli"
)

const (
	webPortFlagName    = "webPort"
	webPortFlagDefault = 12500
	portFlagName       = "port"
	portFlagDefault    = 12501
)

var WebCommand = cli.Command{
	Name:        "web",
	Description: "Starts the server",
	Flags: []cli.Flag{
		cli.IntFlag{Name: webPortFlagName, Value: webPortFlagDefault},
		cli.IntFlag{Name: portFlagName, Value: portFlagDefault},
	},
	Action: runWeb,
}

type webArgs struct {
	webPort int
	port    int
}

func runWeb(ctx *cli.Context) {
	var args webArgs
	args.webPort = ctx.Int(webPortFlagName)
	args.port = ctx.Int(portFlagName)

	startWebServer(&args)
}

func startWebServer(args *webArgs) {
	http.HandleFunc("/", RootHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", args.webPort), nil)
}

func RootHandler(w http.ResponseWriter, _ *http.Request) {
	io.WriteString(w, "root\n")
}
