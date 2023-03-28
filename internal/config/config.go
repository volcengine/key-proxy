/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package config

import "github.com/volcengine/key-proxy/common"

var (
	Conf = &common.Config{}
)

func MustInit(conf *common.Config) {
	Conf = conf
}
