/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package huawei

import (
	"context"
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"net/http"
)

const (
	huaweiSignatureKey = "Authorization"
	vendorName         = "huawei"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &huaweiProvider{Credentials: credential}
	})
}

type huaweiProvider struct {
	Credentials common.Credentials
}

func (s *huaweiProvider) String() string {
	return vendorName
}

func (s *huaweiProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	requestSign := req.Header.Get(huaweiSignatureKey)
	req.Header.Del(huaweiSignatureKey)
	computedSign, err := Sign(s.Credentials.Proxy.AccessKey, s.Credentials.Proxy.SecretKey, req)
	if err != nil {
		return ctx, false, fmt.Errorf("compute signature failed: %v", err)
	}
	return ctx, computedSign == requestSign, nil
}

func (s *huaweiProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	computedSign, err := Sign(s.Credentials.Real.AccessKey, s.Credentials.Real.SecretKey, req)
	if err != nil {
		return fmt.Errorf("compute signature failed: %v", err)
	}
	req.Header.Set(huaweiSignatureKey, computedSign)
	return nil
}
