/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package proxy

import (
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"gopkg.in/yaml.v2"
	"os"
)

func LoadYamlConfig(configFilePath string) (*common.Config, error) {
	var conf common.Config
	if err := unmarshalConfDir(configFilePath, &conf); err != nil {
		return nil, fmt.Errorf("read config failed: %v", err)
	}
	return &conf, nil
}

func fileExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func unmarshalConfDir(path string, target interface{}) error {
	if !fileExist(path) {
		return fmt.Errorf("config file is not existed: %v", path)
	}
	fileBody, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(fileBody, target); err != nil {
		return err
	}
	return nil
}
