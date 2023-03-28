/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package volcengine

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/volcengine/key-proxy/internal/utils"
	"net/http"
	"strings"
	"time"
)

var (
	signedHeadersStr = strings.Join([]string{"content-type", "host", "x-content-sha256", "x-date"}, ";")
	contentType      = "application/json"
)

func hmacSHA256(key []byte, content string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

func hashSHA256(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}

type requestParam struct {
	Body   []byte
	Method string
	Date   time.Time
	Path   string
	Host   string
}

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	Service         string
	Region          string
}

type signRequest struct {
	XDate          string
	Host           string
	ContentType    string
	XContentSha256 string
	Authorization  string
}

func sign(request *http.Request, credential Credentials, signTime time.Time) (signRequest, error) {
	body, err := utils.CopyRequestBody(request)
	if err != nil {
		return signRequest{}, nil
	}
	reqParam := requestParam{
		Body:   body,
		Host:   request.URL.Host,
		Path:   request.URL.Path,
		Method: request.Method,
		Date:   signTime,
	}

	xDate := reqParam.Date.Format("20060102T150405Z")
	shortXDate := xDate[:8]
	XContentSha256 := hashSHA256(reqParam.Body)
	signResult := signRequest{
		Host:           reqParam.Host,  // 设置Host
		XContentSha256: XContentSha256, // 加密body
		XDate:          xDate,          // 设置标准化时间
		ContentType:    contentType,    // 设置Content-Type 为 application/json
	}

	canonicalRequestStr := strings.Join([]string{
		reqParam.Method,
		reqParam.Path,
		request.URL.RawQuery,
		strings.Join([]string{"content-type:" + contentType, "host:" + reqParam.Host, "x-content-sha256:" + XContentSha256, "x-date:" + xDate}, "\n"),
		"",
		signedHeadersStr,
		XContentSha256,
	}, "\n")
	hashedCanonicalRequest := hashSHA256([]byte(canonicalRequestStr))

	credentialScope := strings.Join([]string{shortXDate, credential.Region, credential.Service, "request"}, "/")
	stringToSign := strings.Join([]string{
		"HMAC-SHA256",
		xDate,
		credentialScope,
		hashedCanonicalRequest,
	}, "\n")
	kDate := hmacSHA256([]byte(credential.SecretAccessKey), shortXDate)
	kRegion := hmacSHA256(kDate, credential.Region)
	kService := hmacSHA256(kRegion, credential.Service)
	kSigning := hmacSHA256(kService, "request")
	signature := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))
	signResult.Authorization = fmt.Sprintf("HMAC-SHA256 Credential=%s, SignedHeaders=%s, Signature=%s", credential.AccessKeyID+"/"+credentialScope, signedHeadersStr, signature)
	return signResult, nil
}
