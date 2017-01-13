package cmd

import (
	"flag"
	"fmt"
	"io"
	"net/http"
)

const (
	webPortFlagName    = "webPort"
	webPortFlagDefault = 12500
	portFlagName       = "port"
	portFlagDefault    = 12501
)

var WebCommand = Command{
	Name:        "web",
	Description: "Starts the server",
	Flags: []Flag{
		{Name: webPortFlagName, Value: webPortFlagDefault},
		{Name: portFlagName, Value: portFlagDefault},
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
	fls := flag.FlagSet{}
	fls.IntVar(&a.webPort, webPortFlagName, webPortFlagDefault, "")
	fls.IntVar(&a.port, portFlagName, portFlagDefault, "")
	err := fls.Parse(args)
	if err != nil {
		return err
	}

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
	io.WriteString(w, "root\n")
}
