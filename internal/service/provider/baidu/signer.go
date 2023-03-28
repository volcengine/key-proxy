/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package baidu

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	iso8601Format    = "2006-01-02T15:04:05Z"
	expiredSeconds   = 1800
	authorizationKey = "Authorization"
)

var (
	headsToSign = map[string]struct{}{
		"host":           {},
		"content-length": {},
		"content-type":   {},
		"content-md5":    {},
	}
)

func formatISO8601Date(seconds int64) string {
	tm := time.Unix(seconds, 0).UTC()
	return tm.Format(iso8601Format)
}

func hmacSha256Hex(key, data string) string {
	hasher := hmac.New(sha256.New, []byte(key))
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getCanonicalURIPath(path string) string {
	if len(path) == 0 {
		return "/"
	}
	canonicalPath := path
	if strings.HasPrefix(path, "/") {
		canonicalPath = path[1:]
	}
	canonicalPath = UriEncode(canonicalPath, false)
	return "/" + canonicalPath
}

func getCanonicalQueryString(params url.Values) string {
	if len(params) == 0 {
		return ""
	}

	result := make([]string, 0, len(params))
	for k, v := range params {
		if strings.ToLower(k) == "authorization" {
			continue
		}
		item := UriEncode(k, true) + "="
		if len(v) > 0 {
			item += UriEncode(v[0], true)
		}
		result = append(result, item)
	}
	sort.Strings(result)
	return strings.Join(result, "&")
}

func UriEncode(uri string, encodeSlash bool) string {
	var buf bytes.Buffer
	for _, b := range []byte(uri) {
		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '-' || b == '_' || b == '.' || b == '~' || (b == '/' && !encodeSlash) {
			buf.WriteByte(b)
		} else {
			buf.WriteString(fmt.Sprintf("%%%02X", b))
		}
	}
	return buf.String()
}

func getCanonicalHeaders(headers map[string][]string) (string, []string) {
	canonicalHeaders := make([]string, 0, len(headers))
	signHeaders := make([]string, 0, len(headsToSign))
	for k, v := range headers {
		if len(v) == 0 {
			continue
		}
		headKey := strings.ToLower(k)
		if headKey == "authorization" {
			continue
		}
		_, existed := headsToSign[headKey]
		if existed || (strings.HasPrefix(headKey, "x-bce-") && (headKey != "x-bce-request-id")) {
			headVal := strings.TrimSpace(v[0])
			encoded := UriEncode(headKey, true) + ":" + UriEncode(headVal, true)
			canonicalHeaders = append(canonicalHeaders, encoded)
			signHeaders = append(signHeaders, headKey)
		}
	}
	sort.Strings(canonicalHeaders)
	sort.Strings(signHeaders)
	return strings.Join(canonicalHeaders, "\n"), signHeaders
}

func getSignature(req *http.Request, ak, sk string, timestamp int64) string {
	signDate := formatISO8601Date(timestamp)
	signKeyInfo := fmt.Sprintf("%s/%s/%s/%d", "bce-auth-v1", ak, signDate, expiredSeconds)
	signKey := hmacSha256Hex(sk, signKeyInfo)
	canonicalUri := getCanonicalURIPath(req.URL.Path)
	canonicalQueryString := getCanonicalQueryString(req.URL.Query())
	canonicalHeaders, signedHeadersArr := getCanonicalHeaders(req.Header)

	signedHeaders := ""
	if len(signedHeadersArr) > 0 {
		sort.Strings(signedHeadersArr)
		signedHeaders = strings.Join(signedHeadersArr, ";")
	}

	canonicalParts := []string{req.Method, canonicalUri, canonicalQueryString, canonicalHeaders}
	canonicalReq := strings.Join(canonicalParts, "\n")
	signature := hmacSha256Hex(signKey, canonicalReq)

	authStr := signKeyInfo + "/" + signedHeaders + "/" + signature
	return authStr
}
