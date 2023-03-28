/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package ksyun

import (
	"bytes"
	"context"
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/volcengine/key-proxy/internal/utils"
)

const (
	X_Amz_Date    = "X-Amz-Date"
	Authorization = "Authorization"
	vendorName    = "ksyun"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &ksyunProvider{Credentials: credential}
	})
}

type ksyunProvider struct {
	Credentials common.Credentials
}

func (s *ksyunProvider) String() string {
	return vendorName
}

func (s *ksyunProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	token := req.Header.Get(Authorization)
	err := s.sign(req, s.Credentials.Proxy)
	if err != nil {
		return ctx, false, err
	}
	fakeToken := req.Header.Get(Authorization)
	return ctx, fakeToken == token, nil
}

func (s *ksyunProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	err := s.sign(req, s.Credentials.Real)
	return err
}

func (s *ksyunProvider) sign(req *http.Request, cre common.Credential) error {
	timeStr := req.Header.Get(X_Amz_Date)
	if timeStr == "" {
		return fmt.Errorf("amz date not found in the http header")
	}
	t, _ := time.ParseInLocation("20060102T150405Z", timeStr, time.UTC)
	token := req.Header.Get(Authorization)
	items := strings.Split(token, "/")
	if token == "" || len(items) < 3 {
		return fmt.Errorf("authorization format error")
	}
	region := items[2]

	req.Header.Del(X_Amz_Date)
	req.Header.Del(Authorization)

	fakeSigner := v4.Signer{
		Credentials: credentials.NewStaticCredentials(cre.AccessKey, cre.SecretKey, ""),
	}

	body, err := utils.CopyRequestBody(req)
	if err != nil {
		return fmt.Errorf("copy request body failed: %v", err)
	}

	reader := bytes.NewReader(body)
	seeker := aws.ReadSeekCloser(reader)
	_, err = fakeSigner.Sign(req, seeker, "cdn", region, t)
	if err != nil {
		return err
	}
	return nil
}
