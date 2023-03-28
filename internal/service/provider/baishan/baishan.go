/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package baishan

import (
	"context"
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"net/http"
	"net/url"
)

const (
	tokenKey   = "token"
	vendorName = "baishan"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &baishanProvider{Credentials: credential}
	})
}

type baishanProvider struct {
	Credentials common.Credentials
}

func (s *baishanProvider) String() string {
	return vendorName
}

func (s *baishanProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	rawQuery := req.URL.RawQuery
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return ctx, false, err
	}
	token := query.Get(tokenKey)
	if token == "" {
		return ctx, false, fmt.Errorf("invalid parameters: miss token in the query parameters")
	}
	ctx = context.WithValue(ctx, "rawQuery", query)
	return ctx, s.Credentials.Proxy.AccessToken == token, nil
}

func (s *baishanProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	query := ctx.Value("rawQuery").(url.Values)
	query.Del(tokenKey)
	query.Set(tokenKey, s.Credentials.Real.AccessToken)
	req.URL.RawQuery = query.Encode()
	return nil
}
