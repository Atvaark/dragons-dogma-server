package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/atvaark/dragons-dogma-server/modules/db"
	"github.com/atvaark/dragons-dogma-server/modules/network"
	"github.com/atvaark/dragons-dogma-server/modules/website"
	"github.com/urfave/cli"
)

const (
	webPortFlagName     = "webPort"
	webSteamKeyFlagName = "webSteamKey"
	webRootURLFlagName  = "webRootURL"
	gamePortFlagName    = "gamePort"
	gameCertFileName    = "gameCertFile"
	gameKeyFileName     = "gameKeyFile"
	databaseFileName    = "databaseFile"

	webPortFlagDefault  = 12500
	webSteamKeyDefault  = ""
	webRootURLDefault   = "http://localhost"
	gamePortFlagDefault = 12501
	gameCertFileDefault = "server.crt"
	gameKeyFileDefault  = "server.key"
	databaseFileDefault = "server.db"
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
		cli.StringFlag{Name: databaseFileName, Value: databaseFileDefault},
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
	databaseFile string
}

func (cfg *webConfig) parse(ctx *cli.Context) {
	cfg.webPort = ctx.Int(webPortFlagName)
	cfg.webSteamKey = ctx.String(webSteamKeyFlagName)
	cfg.webRootURL = ctx.String(webRootURLFlagName)
	cfg.gamePort = ctx.Int(gamePortFlagName)
	cfg.gameCertFile = ctx.String(gameCertFileName)
	cfg.gameKeyFile = ctx.String(gameKeyFileName)
	cfg.databaseFile = ctx.String(databaseFileName)

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

	database := startDatabase(&cfg)
	go startGameServer(&cfg, database)
	go startWebServer(&cfg, database)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, os.Kill)
	for range signalChannel {
		// TODO: gracefully shutdown game- and webserver

		err := database.Close()
		if err != nil {
			fmt.Println("failed to close database: ", err)
		}

		return
	}
}

func startDatabase(cfg *webConfig) db.Database {
	database, err := db.NewDatabase(cfg.databaseFile)
	if err != nil {
		panic(err)
	}

	return database
}

func startGameServer(cfg *webConfig, database db.Database) {
	srvConfig := network.ServerConfig{
		Port:     cfg.gamePort,
		CertFile: cfg.gameCertFile,
		KeyFile:  cfg.gameKeyFile,
	}

	srv, err := network.NewServer(srvConfig, database)
	if err != nil {
		panic(err)
	}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func startWebServer(cfg *webConfig, database db.Database) {
	srvConfig := website.WebsiteConfig{
		RootURL: cfg.webRootURL,
		Port:    cfg.webPort,
		AuthConfig: website.AuthConfig{
			SteamKey: cfg.webSteamKey,
		},
	}

	srv := website.NewWebsite(srvConfig, database)

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
