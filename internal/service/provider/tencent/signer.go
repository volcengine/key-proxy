/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package tencent

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/volcengine/key-proxy/internal/utils"
	"net/http"
	"time"
)

const (
	algorithmName = "TC3-HMAC-SHA256"
)

func sha256hex(s []byte) string {
	b := sha256.Sum256(s)
	return hex.EncodeToString(b[:])
}

func hmacSha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

func Sign(req *http.Request, signTime time.Time, ak, sk, host, service string) (string, error) {
	canonicalHeaders := "content-type:application/json\nhost:" + host + "\n"
	signedHeaders := "content-type;host"

	payload, err := utils.CopyRequestBody(req)
	if err != nil {
		return "", err
	}
	hashedRequestPayload := sha256hex(payload)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		req.URL.Path,
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)

	timestamp := signTime.Unix()
	date := signTime.Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	hashedCanonicalRequest := sha256hex([]byte(canonicalRequest))
	string2sign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithmName,
		timestamp,
		credentialScope,
		hashedCanonicalRequest)

	secretDate := hmacSha256(date, "TC3"+sk)
	secretService := hmacSha256(service, secretDate)
	secretSigning := hmacSha256("tc3_request", secretService)
	signature := hex.EncodeToString([]byte(hmacSha256(string2sign, secretSigning)))

	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithmName,
		ak,
		credentialScope,
		signedHeaders,
		signature)

	return authorization, nil
}
