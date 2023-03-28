/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package ucloud

import (
	"bytes"
	"context"
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"io/ioutil"
	"net/http"

	"github.com/volcengine/key-proxy/internal/utils"
)

const (
	ucloudSignatureKey = "Signature"
	ucloudPublicKey    = "PublicKey"
	vendorName         = "ucloud"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &ucloudProvider{Credentials: credential}
	})
}

type ucloudProvider struct {
	Credentials common.Credentials
}

func (s *ucloudProvider) String() string {
	return vendorName
}

func (s *ucloudProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	payload, err := getBody(req)
	if err != nil {
		return ctx, false, err
	}
	fakeSignature, ok := payload[ucloudSignatureKey]
	if !ok {
		return nil, false, fmt.Errorf("signature not found")
	}
	delete(payload, ucloudSignatureKey)
	fakeCre := Credential{
		Ak: s.Credentials.Proxy.AccessKey,
		Sk: s.Credentials.Proxy.SecretKey,
	}
	computeSign := fakeCre.CreateSign(mapToQuery(payload))
	return ctx, fakeSignature == computeSign, nil
}

func (s *ucloudProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	payload, err := getBody(req)
	if err != nil {
		return err
	}
	delete(payload, ucloudSignatureKey)
	realCre := Credential{
		Ak: s.Credentials.Real.AccessKey,
		Sk: s.Credentials.Real.SecretKey,
	}
	payload = realCre.Apply(payload)
	setBody(req, mapToQuery(payload))
	return nil
}
func getBody(req *http.Request) (map[string]string, error) {
	query, err := utils.CopyRequestBody(req)
	if err != nil {
		return nil, err
	}
	payload := queryToMap(string(query))
	return payload, nil
}

func setBody(req *http.Request, body string) {
	length := len(body)
	req.ContentLength = int64(length)
	req.Body = ioutil.NopCloser(bytes.NewReader([]byte(body)))
}
