/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package main

import (
	"flag"
	"github.com/volcengine/key-proxy/pkg/proxy"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "conf-file", "./config.yml", "config file path")
	flag.Parse()

	config, err := proxy.LoadYamlConfig(configFile)
	if err != nil {
		panic(err)
	}
	keyProxy, err := proxy.New(config)
	if err != nil {
		panic(err)
	}
	err = keyProxy.Run()
	if err != nil {
		panic(err)
	}
}
