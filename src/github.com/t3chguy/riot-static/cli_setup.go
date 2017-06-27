// Copyright 2017 Michael Telatynski <7t3chguy@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/matrix-org/gomatrix"
	"io/ioutil"
	"os"
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
