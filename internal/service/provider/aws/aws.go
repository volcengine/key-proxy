/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package aws

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
	authorizationHeader = "Authorization"
	signTimeHeader      = "X-Amz-Date"
	awsRegionKey        = "AwsRegion"
	awsTimeKey          = "AwsTime"
	awsService          = "AwsService"
	vendorName          = "aws"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &awsProvider{Credentials: credential}
	})
}

type awsProvider struct {
	Credentials common.Credentials
}

func (s *awsProvider) String() string {
	return vendorName
}

func (s *awsProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	requestSign := req.Header.Get(authorizationHeader)
	signTimeStr := req.Header.Get(signTimeHeader)
	req.Header.Del(authorizationHeader)
	req.Header.Del(signTimeHeader)
	signTime, err := time.ParseInLocation("20060102T150405Z", signTimeStr, time.UTC)
	if err != nil {
		return ctx, false, fmt.Errorf("parse signing time failed: %v", err)
	}
	items := strings.Split(requestSign, "/")
	if len(items) < 4 {
		return ctx, false, errors.New("authorization format is wrong")
	}
	region := items[2]
	service := items[3]
	ctx = context.WithValue(ctx, awsTimeKey, signTime)
	ctx = context.WithValue(ctx, awsService, service)
	ctx = context.WithValue(ctx, awsRegionKey, region)
	cre := s.Credentials.Proxy
	err = signRequest(ctx, req, cre.AccessKey, cre.SecretKey, cre.AccessToken, region, service, signTime)
	if err != nil {
		return ctx, false, fmt.Errorf("sign aws request failed: %v", err)
	}
	computedSign := req.Header.Get(authorizationHeader)
	return ctx, computedSign == requestSign, nil
}

func (s *awsProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	req.Header.Del(authorizationHeader)
	req.Header.Del(signTimeHeader)
	region := ctx.Value(awsRegionKey).(string)
	service := ctx.Value(awsService).(string)
	signTime := ctx.Value(awsTimeKey).(time.Time)
	cre := s.Credentials.Real
	err := signRequest(ctx, req, cre.AccessKey, cre.SecretKey, cre.AccessToken, region, service, signTime)
	if err != nil {
		return fmt.Errorf("sign aws request failed: %v", err)
	}
	return nil
}
