package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/atvaark/dragons-dogma-server/modules/api"
	"github.com/urfave/cli"
)

const (
	apiPortFlagName            = "port"
	apiServerHostFlagName      = "serverHost"
	apiServerPortFlagName      = "serverPort"
	apiUserFlagName            = "user"
	apiUserTokenFlagName       = "token"
	apiUserTokenFormatFlagName = "tokenFormat"

	apiPortFlagDefault = 12502

	apiServerHostFlagDefault      = "dune.dragonsdogma.com"
	apiServerPortFlagDefault      = 12501
	apiUserTokenFormatFlagDefault = "base64"
)

var ApiCommand = cli.Command{
	Name:        "api",
	Description: "Hosts a server that exposes a JSON endpoint for the ur dragon status .",
	Flags: []cli.Flag{
		cli.IntFlag{Name: apiPortFlagName, Value: apiPortFlagDefault},
		cli.StringFlag{Name: apiServerHostFlagName, Value: apiServerHostFlagDefault},
		cli.IntFlag{Name: apiServerPortFlagName, Value: apiServerPortFlagDefault},
		cli.StringFlag{Name: apiUserFlagName},
		cli.StringFlag{Name: apiUserTokenFlagName},
		cli.StringFlag{Name: apiUserTokenFormatFlagName, Value: apiUserTokenFormatFlagDefault},
	},
	Action: runApi,
}

func parseConfig(ctx *cli.Context) (api.DragonAPIConfig, error) {
	var cfg api.DragonAPIConfig
	cfg.Port = ctx.Int(apiPortFlagName)
	cfg.ServerHost = ctx.String(apiServerHostFlagName)
	cfg.ServerPort = ctx.Int(apiServerPortFlagName)
	cfg.User = ctx.String(apiUserFlagName)

	userTokenArg := ctx.String(apiUserTokenFlagName)
	var userToken []byte
	if len(userTokenArg) == 0 {
		userToken = make([]byte, 0)
	} else {
		userTokenFormat := ctx.String(apiUserTokenFormatFlagName)
		if len(userTokenFormat) == 0 {
			return cfg, errors.New("missing token format")
		}

		switch userTokenFormat {
		case "base64":
			var err error
			userToken, err = base64.StdEncoding.DecodeString(userTokenArg)
			if err != nil {
				return cfg, fmt.Errorf("invalid base64 token: %v", err)
			}

		default:
			return cfg, fmt.Errorf("unknown token format %s", userTokenFormat)
		}
	}

	cfg.UserToken = userToken
	return cfg, nil
}

func runApi(ctx *cli.Context) {
	cfg, err := parseConfig(ctx)
	if err != nil {
		panic(err)
	}

	dragonAPI := api.NewDragonAPI(cfg)
	err = dragonAPI.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
