/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package jingdong

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"github.com/volcengine/key-proxy/internal/utils"
	"net/http"
	"strings"
	"time"
)

const (
	authorizationHeader = "Authorization"
	regionKey           = "Region"
	signTimeKey         = "SignTime"
	serviceKey          = "Service"
	uuidKey             = "Uuid"
	vendorName          = "jingdong"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &jingdongProvider{Credentials: credential}
	})
}

type jingdongProvider struct {
	Credentials common.Credentials
}

func (s *jingdongProvider) String() string {
	return vendorName
}

func (s *jingdongProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	requestSign := req.Header.Get(authorizationHeader)
	signTimeStr := req.Header.Get(signTimeHeaderKey)
	uuidStr := req.Header.Get(uuidHeaderKey)
	req.Header.Del(authorizationHeader)
	req.Header.Del(signTimeHeaderKey)
	req.Header.Del(uuidHeaderKey)
	signTime, err := time.ParseInLocation(timeFormat, signTimeStr, time.UTC)
	if err != nil {
		return ctx, false, fmt.Errorf("parse signing time failed: %v", err)
	}
	items := strings.Split(requestSign, "/")
	if len(items) < 4 {
		return ctx, false, errors.New("authorization format is wrong")
	}
	region := items[2]
	service := items[3]
	ctx = context.WithValue(ctx, signTimeKey, signTime)
	ctx = context.WithValue(ctx, serviceKey, service)
	ctx = context.WithValue(ctx, regionKey, region)
	ctx = context.WithValue(ctx, uuidKey, uuidStr)
	cre := s.Credentials.Proxy

	signer := NewSigner(Credential{
		Ak: cre.AccessKey,
		Sk: cre.SecretKey,
	})
	body, err := utils.CopyRequestBody(req)
	if err != nil {
		return ctx, false, err
	}
	err = signer.signRequest(req, bytes.NewReader(body), service, region, signTime, uuidStr)
	if err != nil {
		return ctx, false, fmt.Errorf("sign jingdong request failed: %v", err)
	}
	computedSign := req.Header.Get(authorizationHeader)
	return ctx, computedSign == requestSign, nil
}

func (s *jingdongProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	signTime := ctx.Value(signTimeKey).(time.Time)
	service := ctx.Value(serviceKey).(string)
	region := ctx.Value(regionKey).(string)
	uuidStr := ctx.Value(uuidKey).(string)
	cre := s.Credentials.Real
	signer := NewSigner(Credential{
		Ak: cre.AccessKey,
		Sk: cre.SecretKey,
	})
	body, err := utils.CopyRequestBody(req)
	if err != nil {
		return err
	}
	err = signer.signRequest(req, bytes.NewReader(body), service, region, signTime, uuidStr)
	if err != nil {
		return fmt.Errorf("sign jingdong request failed: %v", err)
	}
	return nil
}
