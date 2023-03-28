/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package qiniu

import (
	"context"
	"errors"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"net/http"
	"strings"
)

const (
	signatureHeaderKey = "Authorization"
	vendorName         = "qiniu"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &qiniuProvider{Credentials: credential}
	})
}

type qiniuProvider struct {
	Credentials common.Credentials
}

func (s *qiniuProvider) String() string {
	return vendorName
}

func (s *qiniuProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	requestSign := req.Header.Get(signatureHeaderKey)
	if !strings.HasPrefix(requestSign, "QBox ") {
		return ctx, false, errors.New("authorization format is invalid")
	}
	req.Header.Del(signatureHeaderKey)
	requestSign = requestSign[5:]
	computedSign, err := getSignature(req, s.Credentials.Proxy.AccessKey, s.Credentials.Proxy.SecretKey)
	if err != nil {
		return ctx, false, err
	}
	return ctx, requestSign == computedSign, nil
}

func (s *qiniuProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	sign, err := getSignature(req, s.Credentials.Real.AccessKey, s.Credentials.Real.SecretKey)
	if err != nil {
		return err
	}
	req.Header.Set(signatureHeaderKey, "QBox "+sign)
	return nil
}
