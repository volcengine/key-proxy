/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package baidu

import (
	"context"
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"net/http"
	"time"
)

const (
	signTimeKey = "signTimeInt64"
	vendorName  = "baidu"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &baiduProvider{Credentials: credential}
	})
}

type baiduProvider struct {
	Credentials common.Credentials
}

func (s *baiduProvider) String() string {
	return vendorName
}

func (s *baiduProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	requestSign := req.Header.Get(authorizationKey)
	req.Header.Del(authorizationKey)
	signTimeStr := req.Header.Get("x-bce-date")
	signTime, err := time.ParseInLocation(iso8601Format, signTimeStr, time.UTC)
	if err != nil {
		return ctx, false, fmt.Errorf("parse sign time failed, err: %w", err)
	}
	computedSign := getSignature(req, s.Credentials.Proxy.AccessKey, s.Credentials.Proxy.SecretKey, signTime.Unix())
	ctx = context.WithValue(ctx, signTimeKey, signTime.Unix())
	return ctx, requestSign == computedSign, nil
}

func (s *baiduProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	signTimeUnix := ctx.Value(signTimeKey).(int64)
	signature := getSignature(req, s.Credentials.Real.AccessKey, s.Credentials.Real.SecretKey, signTimeUnix)
	req.Header.Set(authorizationKey, signature)
	return nil
}
