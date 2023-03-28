/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package tencent

import (
	"context"
	"errors"
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	signatureHeaderKey = "Authorization"
	timestampHeaderKey = "X-TC-Timestamp"
	hostHeaderKey      = "Host"
	signTimeKey        = "VolcTime"
	serviceKey         = "VolcService"
	regionKey          = "VolcRegion"
	vendorName         = "tencent"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &tencentProvider{Credentials: credential}
	})
}

type tencentProvider struct {
	Credentials common.Credentials
}

func (s *tencentProvider) String() string {
	return vendorName
}

func (s *tencentProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	requestSign := req.Header.Get(signatureHeaderKey)
	req.Header.Del(signatureHeaderKey)
	signTimestampKey := req.Header.Get(timestampHeaderKey)
	signTimestamp, err := strconv.ParseInt(signTimestampKey, 10, 64)
	if err != nil {
		return ctx, false, fmt.Errorf("parse signing time failed: %v", err)
	}
	signTime := time.Unix(signTimestamp, 0).UTC()
	items := strings.Split(requestSign, "/")
	if len(items) < 3 {
		return ctx, false, errors.New("authorization format is wrong")
	}
	service := items[2]
	ctx = context.WithValue(ctx, signTimeKey, signTime)
	ctx = context.WithValue(ctx, serviceKey, service)
	cre := s.Credentials.Proxy
	computedSign, err := Sign(req, signTime, cre.AccessKey, cre.SecretKey, req.URL.Host, service)
	if err != nil {
		return ctx, false, fmt.Errorf("compute signature failed: %v", err)
	}
	return ctx, computedSign == requestSign, nil
}

func (s *tencentProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	service := ctx.Value(serviceKey).(string)
	signTime := ctx.Value(signTimeKey).(time.Time)
	cre := s.Credentials.Real
	computedSign, err := Sign(req, signTime, cre.AccessKey, cre.SecretKey, req.URL.Host, service)
	if err != nil {
		return fmt.Errorf("compute signature failed: %v", err)
	}
	req.Header.Set(signatureHeaderKey, computedSign)
	return nil
}
