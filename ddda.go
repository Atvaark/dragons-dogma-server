package main

import (
	"os"
	"log"


	"github.com/urfave/cli"

	"github.com/atvaark/dragons-dogma-server/cmd"
)

const AppVersion = "0.0.1"

func main() {
	
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	
	user := os.Getenv("USER")
	if user == "" {
		log.Fatal("$USER must be set")
	}
	
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("$TOKEN must be set")
	}
	
	args := []string {
		os.Args[0],
		"api",
		"--port",
		port,
		"--user",
		user,
		"--token",
		token,
	}



	app := cli.NewApp()
	app.Name = "Dragon's Dogma Server"
	app.Version = AppVersion
	app.Commands = []cli.Command{
		cmd.WebCommand,
		cmd.TestCommand,
		cmd.ApiCommand,
	}

	app.Run(args)
}
