/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package akamai

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/volcengine/key-proxy/internal/utils"
	"net/http"
	"sort"
	"strings"
	"unicode"
)

var (
	HeaderToSign []string
	MaxBody      = 131072 // 128K
)

func stringMinifier(in string) (out string) {
	white := false
	for _, c := range in {
		if unicode.IsSpace(c) {
			if !white {
				out = out + " "
			}
			white = true
		} else {
			out = out + string(c)
			white = false
		}
	}
	return
}

func concatPathQuery(path, query string) string {
	if query == "" {
		return path
	}
	return path + "?" + query
}

func createSignature(message string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func createHash(data string) string {
	h := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(h[:])
}

func canonicalizeHeaders(req *http.Request) string {
	unsortedHeader := make([]string, 0, len(req.Header))
	sortedHeader := make([]string, 0, len(req.Header))
	for k := range req.Header {
		unsortedHeader = append(unsortedHeader, k)
	}
	sort.Strings(unsortedHeader)
	for _, k := range unsortedHeader {
		for _, sign := range HeaderToSign {
			if sign == k {
				v := strings.TrimSpace(req.Header.Get(k))
				sortedHeader = append(sortedHeader, strings.ToLower(k)+":"+strings.ToLower(stringMinifier(v)))
			}
		}
	}
	return strings.Join(sortedHeader, "\t")

}

func signingKey(clientSecret string, timestamp string) string {
	return createSignature(timestamp, clientSecret)
}

func createContentHash(req *http.Request) string {
	var contentHash, preparedBody string
	if req.Body != nil {
		b, _ := utils.CopyRequestBody(req)
		preparedBody = string(b)
	}

	if req.Method == "POST" && len(preparedBody) > 0 {
		if len(preparedBody) > MaxBody {
			preparedBody = preparedBody[0:MaxBody]
		}
		contentHash = createHash(preparedBody)
	}
	return contentHash
}

func signingData(req *http.Request, authHeader string) string {
	dataSign := []string{
		req.Method,
		req.URL.Scheme,
		req.URL.Host,
		concatPathQuery(req.URL.EscapedPath(), req.URL.RawQuery),
		canonicalizeHeaders(req),
		createContentHash(req),
		authHeader,
	}
	return strings.Join(dataSign, "\t")
}

func signingRequest(clientSecret string, req *http.Request, authHeader string, timestamp string) string {
	return createSignature(signingData(req, authHeader), signingKey(clientSecret, timestamp))
}

func createAuthHeader(req *http.Request, clientToken, accessToken, clientSecret string, timestamp string, nonce string) string {
	authHeader := fmt.Sprintf("EG1-HMAC-SHA256 client_token=%s;access_token=%s;timestamp=%s;nonce=%s;",
		clientToken,
		accessToken,
		timestamp,
		nonce,
	)
	return authHeader + "signature=" + signingRequest(clientSecret, req, authHeader, timestamp)
}
