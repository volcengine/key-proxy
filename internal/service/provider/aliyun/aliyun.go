/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package aliyun

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/base"
	"github.com/volcengine/key-proxy/internal/service/provider"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const (
	aliyunSignatureKey   = "Signature"
	aliyunAccessKeIdyKey = "AccessKeyId"
	vendorName           = "aliyun"
)

func init() {
	provider.RegisterProvider(vendorName, func(credential common.Credentials) provider.IProvider {
		return &aliyunProvider{Credentials: credential}
	})
}

type aliyunProvider struct {
	Credentials common.Credentials
}

func (s *aliyunProvider) String() string {
	return vendorName
}

func (s *aliyunProvider) ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error) {
	q := req.URL.Query()
	q.Set(aliyunAccessKeIdyKey, s.Credentials.Proxy.AccessKey)
	requestSign := q.Get(aliyunSignatureKey)
	q.Del(aliyunSignatureKey)
	req.URL.RawQuery = base.QuickEncode(q)
	computedSign := s.sign(req.Method, q, s.Credentials.Proxy.SecretKey)
	return ctx, computedSign == requestSign, nil
}

func (s *aliyunProvider) ResignRequest(ctx context.Context, req *http.Request) error {
	q := req.URL.Query()
	q.Set(aliyunAccessKeIdyKey, s.Credentials.Real.AccessKey)
	q.Set(aliyunSignatureKey, s.sign(req.Method, q, s.Credentials.Real.SecretKey))
	req.URL.RawQuery = base.QuickEncode(q)
	return nil
}

func (s *aliyunProvider) sign(method string, q url.Values, secret string) string {
	var queryKeys []string
	for key := range q {
		queryKeys = append(queryKeys, key)
	}
	sort.Strings(queryKeys)
	var strBuilder strings.Builder
	for _, key := range queryKeys {
		strBuilder.WriteString(percentEncode(key))
		strBuilder.WriteString("=")
		strBuilder.WriteString(percentEncode(q.Get(key)))
		strBuilder.WriteString("&")
	}
	queryParts := strBuilder.String()
	strBuilder.Reset()
	strBuilder.WriteString(method)
	strBuilder.WriteString("&")
	strBuilder.WriteString(percentEncode("/"))
	strBuilder.WriteString("&")
	strBuilder.WriteString(percentEncode(queryParts[:len(queryParts)-1]))
	str2Sign := strBuilder.String()
	return shaHmac1(str2Sign, secret+"&")
}

func shaHmac1(sign string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha1.New, key)
	_, _ = h.Write([]byte(sign))
	sb := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(sb)
}

func percentEncode(s string) string {
	v := url.QueryEscape(s)
	v = strings.ReplaceAll(v, "+", "%20")
	v = strings.ReplaceAll(v, "*", "%2A")
	v = strings.ReplaceAll(v, "%7E", "~")
	return v
}
