/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package wangsu

import (
	"context"
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"net/http"
)

const (
	dateHeaderKey       = "Date"
	authorizationKey    = "Authorization"
	authorizationPrefix = "Basic "
	vendorName          = "wangsu"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &wangsuProvider{Credentials: credential}
	})
}

type wangsuProvider struct {
	Credentials common.Credentials
}

func (s *wangsuProvider) String() string {
	return vendorName
}

func (s *wangsuProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	date := req.Header.Get(dateHeaderKey)
	if date == "" {
		return ctx, false, fmt.Errorf("invalid parameters: miss Date in the query parameters")
	}
	ctx = context.WithValue(ctx, "Date", date)
	fakeSignature := req.Header.Get(authorizationKey)
	computeSignature := authorizationPrefix + authorize(s.Credentials.Proxy.AccessKey, hmac64(date, s.Credentials.Proxy.SecretKey))
	return ctx, fakeSignature == computeSignature, nil
}

func (s *wangsuProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	date := ctx.Value("Date").(string)
	req.Header.Del(authorizationKey)
	realSignature := authorizationPrefix + authorize(s.Credentials.Real.AccessKey, hmac64(date, s.Credentials.Real.SecretKey))
	req.Header.Set(authorizationKey, realSignature)
	return nil
}
