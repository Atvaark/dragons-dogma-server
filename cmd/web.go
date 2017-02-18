package cmd

import (
	"fmt"
	"log"
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

	log.Println("Starting")
	database := startDatabase(&cfg)
	gameServer := startGameServer(&cfg, database)
	gameWebsite := startGameWebsite(&cfg, database)
	log.Println("Started")

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	for range signalChannel {
		log.Println("Stopping")
		var err error
		err = gameWebsite.Close()
		if err != nil {
			log.Println("failed to close website: ", err)
		}

		err = gameServer.Close()
		if err != nil {
			log.Println("failed to close server: ", err)
		}

		err = database.Close()
		if err != nil {
			log.Println("failed to close database: ", err)
		}

		log.Println("Stopped")

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

func startGameServer(cfg *webConfig, database db.Database) *network.Server {
	srvConfig := network.ServerConfig{
		Port:     cfg.gamePort,
		CertFile: cfg.gameCertFile,
		KeyFile:  cfg.gameKeyFile,
	}

	srv, err := network.NewServer(srvConfig, database)
	if err != nil {
		panic(err)
	}

	go func() {
		err = srv.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	return srv
}

func startGameWebsite(cfg *webConfig, database db.Database) *website.Website {
	srvConfig := website.WebsiteConfig{
		RootURL: cfg.webRootURL,
		Port:    cfg.webPort,
		AuthConfig: website.AuthConfig{
			SteamKey: cfg.webSteamKey,
		},
	}

	gameWebsite := website.NewWebsite(srvConfig, database)

	go func() {
		err := gameWebsite.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	return gameWebsite
}
