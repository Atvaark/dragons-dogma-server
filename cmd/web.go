package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/atvaark/dragons-dogma-server/modules/network"
	"github.com/urfave/cli"
)

const (
	webPortFlagName  = "webPort"
	gamePortFlagName = "gamePort"
	gameCertFileName = "gameCertFile"
	gameKeyFileName  = "gameKeyFile"

	webPortFlagDefault  = 12500
	gamePortFlagDefault = 12501
	gameCertFileDefault = "server.crt"
	gameKeyFileDefault  = "server.key"
)

var WebCommand = cli.Command{
	Name:        "web",
	Description: "Starts the server",
	Flags: []cli.Flag{
		cli.IntFlag{Name: webPortFlagName, Value: webPortFlagDefault},
		cli.IntFlag{Name: gamePortFlagName, Value: gamePortFlagDefault},
		cli.StringFlag{Name: gameCertFileName, Value: gameCertFileDefault},
		cli.StringFlag{Name: gameKeyFileName, Value: gameKeyFileDefault},
	},
	Action: runWeb,
}

type webConfig struct {
	webPort      int
	gamePort     int
	gameCertFile string
	gameKeyFile  string
}

func (cfg *webConfig) parse(ctx *cli.Context) {
	cfg.webPort = ctx.Int(webPortFlagName)
	cfg.gamePort = ctx.Int(gamePortFlagName)
	cfg.gameCertFile = ctx.String(gameCertFileName)
	cfg.gameKeyFile = ctx.String(gameKeyFileName)
}

func runWeb(ctx *cli.Context) {
	var cfg webConfig
	cfg.parse(ctx)

	go startGameServer(&cfg)
	startWebServer(&cfg)
}

func startGameServer(cfg *webConfig) {
	srvConfig := network.ServerConfig{
		Port:     cfg.gamePort,
		CertFile: cfg.gameCertFile,
		KeyFile:  cfg.gameKeyFile,
	}

	srv, err := network.NewServer(srvConfig)
	if err != nil {
		panic(err)
	}

	srv.ListenAndServe()
}

func startWebServer(cfg *webConfig) {
	http.HandleFunc("/", RootHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.webPort), nil)
	if err != nil {
		panic(err)
	}
}

func RootHandler(w http.ResponseWriter, _ *http.Request) {
	io.WriteString(w, "root\n")
}
