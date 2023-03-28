/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package service

import (
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"

	_ "github.com/volcengine/key-proxy/internal/service/provider/akamai"
	_ "github.com/volcengine/key-proxy/internal/service/provider/aliyun"
	_ "github.com/volcengine/key-proxy/internal/service/provider/aws"
	_ "github.com/volcengine/key-proxy/internal/service/provider/baidu"
	_ "github.com/volcengine/key-proxy/internal/service/provider/baishan"
	_ "github.com/volcengine/key-proxy/internal/service/provider/huawei"
	_ "github.com/volcengine/key-proxy/internal/service/provider/jingdong"
	_ "github.com/volcengine/key-proxy/internal/service/provider/ksyun"
	_ "github.com/volcengine/key-proxy/internal/service/provider/qiniu"
	_ "github.com/volcengine/key-proxy/internal/service/provider/tencent"
	_ "github.com/volcengine/key-proxy/internal/service/provider/ucloud"
	_ "github.com/volcengine/key-proxy/internal/service/provider/volcengine"
	_ "github.com/volcengine/key-proxy/internal/service/provider/wangsu"
)

var (
	providerService provider.IProviderService
)

func GetProviderService() provider.IProviderService {
	return providerService
}

func Init(config *common.Config) error {
	var err error
	providerService, err = provider.New(config.Endpoints)
	if err != nil {
		return err
	}
	return nil
}
