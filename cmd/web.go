package cmd

import (
	"fmt"
	"io"
	"net/http"
)

const (
	webPortFlagName = "webPort"
	portFlagName    = "port"
)

var WebCommand = Command{
	Name:        "web",
	Description: "Starts the server",
	Flags: []Flag{
		{Name: webPortFlagName, Value: "12500"},
		{Name: portFlagName, Value: "12501"},
	},
	action: runWeb,
	args: func() commandArgs {
		return &webArgs{}
	},
}

type webArgs struct {
	webPort int
	port    int
}

func (a *webArgs) init(flags []Flag, args []string) error {
	// TODO: Read args
	a.webPort = 12500
	a.port = 12501
	return nil
}

func runWeb(ctx *context) {
	fmt.Println("web")
	args := ctx.args.(*webArgs)

	fmt.Println(args)

	http.HandleFunc("/", RootHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", args.webPort), nil)
}

func RootHandler(w http.ResponseWriter, _ *http.Request) {
	io.WriteString(w, "root!\n")
}
