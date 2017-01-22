package cmd

import (
	"fmt"

	"github.com/atvaark/dragons-dogma-server/modules/network"
	"github.com/atvaark/dragons-dogma-server/modules/website"
	"github.com/urfave/cli"
	"strings"
)

const (
	webPortFlagName     = "webPort"
	webSteamKeyFlagName = "webSteamKey"
	webRootURLFlagName  = "webRootURL"
	gamePortFlagName    = "gamePort"
	gameCertFileName    = "gameCertFile"
	gameKeyFileName     = "gameKeyFile"

	webPortFlagDefault  = 12500
	webSteamKeyDefault  = ""
	webRootURLDefault   = "http://localhost"
	gamePortFlagDefault = 12501
	gameCertFileDefault = "server.crt"
	gameKeyFileDefault  = "server.key"
)

var WebCommand = cli.Command{
	Name:        "web",
	Description: "Starts the server",
	Flags: []cli.Flag{
		cli.IntFlag{Name: webPortFlagName, Value: webPortFlagDefault},
		cli.StringFlag{Name: webSteamKeyFlagName, Value: webSteamKeyDefault},
		cli.StringFlag{Name: webRootURLFlagName, Value: webRootURLDefault},
		cli.IntFlag{Name: gamePortFlagName, Value: gamePortFlagDefault},
		cli.StringFlag{Name: gameCertFileName, Value: gameCertFileDefault},
		cli.StringFlag{Name: gameKeyFileName, Value: gameKeyFileDefault},
	},
	Action: runWeb,
}

type webConfig struct {
	webPort      int
	webSteamKey  string
	webRootURL   string
	gamePort     int
	gameCertFile string
	gameKeyFile  string
}

func (cfg *webConfig) parse(ctx *cli.Context) {
	cfg.webPort = ctx.Int(webPortFlagName)
	cfg.webSteamKey = ctx.String(webSteamKeyFlagName)
	cfg.webRootURL = ctx.String(webRootURLFlagName)
	cfg.gamePort = ctx.Int(gamePortFlagName)
	cfg.gameCertFile = ctx.String(gameCertFileName)
	cfg.gameKeyFile = ctx.String(gameKeyFileName)

	if cfg.webRootURL == webRootURLDefault && cfg.webPort != 80 {
		cfg.webRootURL += fmt.Sprintf(":%d", cfg.webPort)
	}

	if !strings.HasSuffix(cfg.webRootURL, "/") {
		cfg.webRootURL += "/"
	}

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
	srvConfig := website.WebsiteConfig{
		RootURL: cfg.webRootURL,
		Port:    cfg.webPort,
		AuthConfig: website.AuthConfig{
			SteamKey: cfg.webSteamKey,
		},
	}

	srv := website.NewWebsite(srvConfig)
	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
