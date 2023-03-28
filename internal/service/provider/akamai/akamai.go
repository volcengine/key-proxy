/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package akamai

import (
	"context"
	"errors"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"net/http"
	"regexp"
)

const (
	signatureKey       = "Authorization"
	akamaiTimestampKey = "AkamaiTimestamp"
	akamaiNonceKey     = "AkamaiNonce"
	vendorName         = "akamai"
)

var timestampNonceRe = regexp.MustCompile(`timestamp=(.*?);nonce=(.*?);`)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &akamaiProvider{Credentials: credential}
	})
}

type akamaiProvider struct {
	Credentials common.Credentials
}

func (s *akamaiProvider) String() string {
	return vendorName
}

func (s *akamaiProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	requestSign := req.Header.Get(signatureKey)
	matches := timestampNonceRe.FindStringSubmatch(requestSign)
	if len(matches) != 3 {
		return ctx, false, errors.New("wrong authorization format")
	}
	cre := s.Credentials.Proxy
	computedSign := createAuthHeader(req, cre.ClientToken, cre.AccessToken, cre.ClientSecret, matches[1], matches[2])
	ctx = context.WithValue(ctx, akamaiTimestampKey, matches[1])
	ctx = context.WithValue(ctx, akamaiNonceKey, matches[2])
	return ctx, computedSign == requestSign, nil
}

func (s *akamaiProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	cre := s.Credentials.Real
	timestamp := ctx.Value(akamaiTimestampKey).(string)
	nonce := ctx.Value(akamaiNonceKey).(string)
	req.Header.Set(signatureKey, createAuthHeader(req, cre.ClientToken, cre.AccessToken, cre.ClientSecret, timestamp, nonce))
	return nil
}
