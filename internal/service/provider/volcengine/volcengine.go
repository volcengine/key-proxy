/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package volcengine

import (
	"context"
	"errors"
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"net/http"
	"strings"
	"time"
)

const (
	signatureHeaderKey = "Authorization"
	signTimeHeaderKey  = "X-Date"
	signTimeKey        = "VolcTime"
	serviceKey         = "VolcService"
	regionKey          = "VolcRegion"
	vendorName         = "volcengine"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &volcengineProvider{Credentials: credential}
	})
}

type volcengineProvider struct {
	Credentials common.Credentials
}

func (s *volcengineProvider) String() string {
	return vendorName
}

func (s *volcengineProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	requestSign := req.Header.Get(signatureHeaderKey)
	signTimeStr := req.Header.Get(signTimeHeaderKey)
	req.Header.Del(signatureHeaderKey)
	signTime, err := time.ParseInLocation("20060102T150405Z", signTimeStr, time.UTC)
	if err != nil {
		return ctx, false, fmt.Errorf("parse signing time failed: %v", err)
	}
	items := strings.Split(requestSign, "/")
	if len(items) < 3 {
		return ctx, false, errors.New("authorization format is wrong")
	}
	region := items[2]
	service := items[3]
	ctx = context.WithValue(ctx, signTimeKey, signTime)
	ctx = context.WithValue(ctx, serviceKey, service)
	ctx = context.WithValue(ctx, regionKey, region)
	cre := s.Credentials.Proxy
	signResult, err := sign(req, Credentials{AccessKeyID: cre.AccessKey, SecretAccessKey: cre.SecretKey, Service: service, Region: region}, signTime)
	if err != nil {
		return ctx, false, fmt.Errorf("compute signature failed: %v", err)
	}
	return ctx, signResult.Authorization == requestSign, nil
}

func (s *volcengineProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	req.Header.Del(signatureHeaderKey)
	region := ctx.Value(regionKey).(string)
	service := ctx.Value(serviceKey).(string)
	signTime := ctx.Value(signTimeKey).(time.Time)
	cre := s.Credentials.Real
	signResult, err := sign(req, Credentials{AccessKeyID: cre.AccessKey, SecretAccessKey: cre.SecretKey, Service: service, Region: region}, signTime)
	if err != nil {
		return fmt.Errorf("compute signature failed: %v", err)
	}
	req.Header.Set(signatureHeaderKey, signResult.Authorization)
	return nil
}
