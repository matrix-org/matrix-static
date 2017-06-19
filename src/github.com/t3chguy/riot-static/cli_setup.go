package main

import (
	"github.com/matrix-org/gomatrix"
	"os"
	"io/ioutil"
	"fmt"
	"encoding/json"
)

var cli *gomatrix.Client
var config *gomatrix.RespRegister

func setupCli() {
	if _, err := os.Stat("./config.json"); err == nil {
		file, e := ioutil.ReadFile("./config.json")
		if e != nil {
			fmt.Printf("File error: %v\n", e)
			os.Exit(1)
		}

		json.Unmarshal(file, &config)
	}

	if config == nil {
		config = new(gomatrix.RespRegister)
	}

	if config.HomeServer == "" {
		config.HomeServer = "https://matrix.org"
	}

	cli, _ = gomatrix.NewClient(config.HomeServer, "", "")

	if config.AccessToken == "" || config.UserID == "" {
		register, inter, err := cli.RegisterGuest(&gomatrix.ReqRegister{})

		if err == nil && inter == nil && register != nil {
			register.HomeServer = config.HomeServer
			config = register
		} else {
			fmt.Println("Error encountered during guest registration")
			os.Exit(1)
		}

		configJson, _ := json.Marshal(config)
		err = ioutil.WriteFile("./config.json", configJson, 0600)
		if err != nil {
			fmt.Println(err)
		}
	}

	cli.SetCredentials(config.UserID, config.AccessToken)
}