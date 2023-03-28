/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package huawei

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/volcengine/key-proxy/internal/utils"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	BasicDateFormat     = "20060102T150405Z"
	Algorithm           = "SDK-HMAC-SHA256"
	HeaderXDate         = "X-Sdk-Date"
	HeaderHost          = "host"
	HeaderContentSha256 = "X-Sdk-Content-Sha256"
)

func hmacSha256(key []byte, data string) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	if _, err := h.Write([]byte(data)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func CanonicalRequest(r *http.Request, signedHeaders []string) (string, error) {
	var encodedHex string
	var err error
	if hex := r.Header.Get(HeaderContentSha256); hex != "" {
		encodedHex = hex
	} else {
		data, err := utils.CopyRequestBody(r)
		if err != nil {
			return "", err
		}
		encodedHex, err = HexEncodeSHA256Hash(data)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		r.Method,
		CanonicalURI(r),
		CanonicalQueryString(r),
		CanonicalHeaders(r, signedHeaders),
		strings.Join(signedHeaders, ";"),
		encodedHex,
	), err
}

func CanonicalURI(r *http.Request) string {
	pattens := strings.Split(r.URL.Path, "/")
	var uri []string
	for _, v := range pattens {
		uri = append(uri, escape(v))
	}
	urlpath := strings.Join(uri, "/")
	if len(urlpath) == 0 || urlpath[len(urlpath)-1] != '/' {
		urlpath = urlpath + "/"
	}
	return urlpath
}

func CanonicalQueryString(r *http.Request) string {
	var keys []string
	query := r.URL.Query()
	for key := range query {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var a []string
	for _, key := range keys {
		k := escape(key)
		sort.Strings(query[key])
		for _, v := range query[key] {
			kv := fmt.Sprintf("%s=%s", k, escape(v))
			a = append(a, kv)
		}
	}
	queryStr := strings.Join(a, "&")
	r.URL.RawQuery = queryStr
	return queryStr
}

func CanonicalHeaders(r *http.Request, signerHeaders []string) string {
	var a []string
	header := make(map[string][]string, len(r.Header))
	for k, v := range r.Header {
		header[strings.ToLower(k)] = v
	}
	for _, key := range signerHeaders {
		value := header[key]
		if strings.EqualFold(key, HeaderHost) {
			value = []string{r.Host}
		}
		sort.Strings(value)
		for _, v := range value {
			a = append(a, key+":"+strings.TrimSpace(v))
		}
	}
	return strings.Join(a, "\n") + "\n"
}

func SignedHeaders(r *http.Request) []string {
	res := make([]string, 0, len(r.Header))
	for key := range r.Header {
		lowerKey := strings.ToLower(key)
		if lowerKey == "content-type" {
			continue
		}
		res = append(res, lowerKey)
	}
	sort.Strings(res)
	return res
}

func StringToSign(canonicalRequest string, t time.Time) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(canonicalRequest))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n%s\n%x",
		Algorithm, t.UTC().Format(BasicDateFormat), hash.Sum(nil)), nil
}

func SignStringToSign(stringToSign string, signingKey []byte) (string, error) {
	hm, err := hmacSha256(signingKey, stringToSign)
	return fmt.Sprintf("%x", hm), err
}

func HexEncodeSHA256Hash(body []byte) (string, error) {
	hash := sha256.New()
	if body == nil {
		body = []byte("")
	}
	_, err := hash.Write(body)
	return fmt.Sprintf("%x", hash.Sum(nil)), err
}

func AuthHeaderValue(signature, accessKey string, signedHeaders []string) string {
	return fmt.Sprintf("%s Access=%s, SignedHeaders=%s, Signature=%s", Algorithm, accessKey, strings.Join(signedHeaders, ";"), signature)
}

func Sign(ak, sk string, r *http.Request) (res string, err error) {
	var t time.Time
	if dt := r.Header.Get(HeaderXDate); dt != "" {
		t, _ = time.Parse(BasicDateFormat, dt)
	}
	signedHeaders := SignedHeaders(r)
	canonicalRequest, err := CanonicalRequest(r, signedHeaders)
	if err != nil {
		return
	}
	stringToSign, err := StringToSign(canonicalRequest, t)
	if err != nil {
		return
	}
	signature, err := SignStringToSign(stringToSign, []byte(sk))
	if err != nil {
		return
	}
	res = AuthHeaderValue(signature, ak, signedHeaders)
	return
}
