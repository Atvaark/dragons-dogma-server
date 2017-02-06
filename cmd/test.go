package cmd

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/atvaark/dragons-dogma-server/modules/network"
	"github.com/urfave/cli"
)

const (
	testHostFlagName            = "host"
	testPortFlagName            = "port"
	testUserFlagName            = "user"
	testUserTokenFlagName       = "token"
	testUserTokenFormatFlagName = "tokenFormat"

	testHostFlagDefault            = "dune.dragonsdogma.com"
	testPortFlagDefault            = 12501
	testUserTokenFormatFlagDefault = "base64"
)

var TestCommand = cli.Command{
	Name:        "test",
	Description: "Tests the server connection by logging in a user",
	Flags: []cli.Flag{
		cli.StringFlag{Name: testHostFlagName, Value: testHostFlagDefault},
		cli.IntFlag{Name: testPortFlagName, Value: testPortFlagDefault},
		cli.StringFlag{Name: testUserFlagName},
		cli.StringFlag{Name: testUserTokenFlagName},
		cli.StringFlag{Name: testUserTokenFormatFlagName, Value: testUserTokenFormatFlagDefault},
	},
	Action: runTest,
}

type testConfig struct {
	host      string
	port      int
	user      string
	userToken []byte
}

func (cfg *testConfig) parse(ctx *cli.Context) error {
	cfg.host = ctx.String(testHostFlagName)
	cfg.port = ctx.Int(testPortFlagName)
	cfg.user = ctx.String(testUserFlagName)

	userTokenArg := ctx.String(testUserTokenFlagName)
	var userToken []byte
	if len(userTokenArg) == 0 {
		userToken = make([]byte, 0)
	} else {
		userTokenFormat := ctx.String(testUserTokenFormatFlagName)
		if len(userTokenFormat) == 0 {
			return errors.New("missing token format")
		}

		switch userTokenFormat {
		case "base64":
			var err error
			userToken, err = base64.StdEncoding.DecodeString(userTokenArg)
			if err != nil {
				return fmt.Errorf("invalid base64 token: %v", err)
			}

		default:
			return fmt.Errorf("unknown token format %s", userTokenFormat)
		}
	}

	cfg.userToken = userToken
	return nil
}

func runTest(ctx *cli.Context) {
	var cfg testConfig
	err := cfg.parse(ctx)
	if err != nil {
		panic(err)
	}

	client := network.NewClient(network.ClientConfig{
		Host:      cfg.host,
		Port:      cfg.port,
		User:      cfg.user,
		UserToken: cfg.userToken,
	})

	err = client.Connect()
	if err != nil {
		panic(err)
	}

	dragon, err := client.GetOnlineUrDragon()
	if err != nil {
		panic(err)
	}

	dragonJson, err := json.MarshalIndent(dragon, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dragonJson))

	err = client.Disconnect()
	if err != nil {
		panic(err)
	}
}
