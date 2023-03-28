/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package jingdong

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	authHeaderPrefix  = "JDCLOUD2-HMAC-SHA256"
	timeFormat        = "20060102T150405Z"
	shortTimeFormat   = "20060102"
	emptyStringSHA256 = `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
	signTimeHeaderKey = "x-jdcloud-date"
	uuidHeaderKey     = "x-jdcloud-nonce"
)

var ignoredHeaders = []string{"Authorization", "User-Agent", "X-Jdcloud-Request-Id"}
var noEscape [256]bool

type Credential struct {
	Ak string
	Sk string
}

func init() {
	for i := 0; i < len(noEscape); i++ {
		noEscape[i] = (i >= 'A' && i <= 'Z') ||
			(i >= 'a' && i <= 'z') ||
			(i >= '0' && i <= '9') ||
			i == '-' ||
			i == '.' ||
			i == '_' ||
			i == '~'
	}
}

type Signer struct {
	Credentials Credential
}

func NewSigner(credsProvider Credential) *Signer {
	return &Signer{
		Credentials: credsProvider,
	}
}

type signingCtx struct {
	ServiceName      string
	Region           string
	Request          *http.Request
	Body             io.ReadSeeker
	Query            url.Values
	Time             time.Time
	SignedHeaderVals http.Header

	credValues         Credential
	formattedTime      string
	formattedShortTime string

	bodyDigest       string
	signedHeaders    string
	canonicalHeaders string
	canonicalString  string
	credentialString string
	stringToSign     string
	signature        string
	authorization    string
}

func (v4 Signer) signRequest(r *http.Request, body io.ReadSeeker, service, region string, signTime time.Time, uuid string) error {
	ctx := &signingCtx{
		Request:     r,
		Body:        body,
		Query:       r.URL.Query(),
		Time:        signTime,
		ServiceName: service,
		Region:      region,
	}

	for key := range ctx.Query {
		sort.Strings(ctx.Query[key])
	}

	if ctx.isRequestSigned() {
		ctx.Time = time.Now()
	}

	ctx.credValues = v4.Credentials
	ctx.build(uuid)

	return nil
}

func (ctx *signingCtx) build(uuid string) {
	ctx.buildTime()
	ctx.Request.Header.Set(uuidHeaderKey, uuid)
	ctx.buildCredentialString()
	ctx.buildBodyDigest()

	unsignedHeaders := ctx.Request.Header
	ctx.buildCanonicalHeaders(unsignedHeaders)
	ctx.buildCanonicalString()
	ctx.buildStringToSign()
	ctx.buildSignature()

	parts := []string{
		authHeaderPrefix + " Credential=" + ctx.credValues.Ak + "/" + ctx.credentialString,
		"SignedHeaders=" + ctx.signedHeaders,
		"Signature=" + ctx.signature,
	}
	ctx.Request.Header.Set("Authorization", strings.Join(parts, ", "))
}

func (ctx *signingCtx) buildTime() {
	ctx.formattedTime = ctx.Time.UTC().Format(timeFormat)
	ctx.formattedShortTime = ctx.Time.UTC().Format(shortTimeFormat)

	ctx.Request.Header.Set(signTimeHeaderKey, ctx.formattedTime)
}

func (ctx *signingCtx) buildNonce(uuid string) {

}

func (ctx *signingCtx) buildCredentialString() {
	ctx.credentialString = strings.Join([]string{
		ctx.formattedShortTime,
		ctx.Region,
		ctx.ServiceName,
		"jdcloud2_request",
	}, "/")
}

func (ctx *signingCtx) buildCanonicalHeaders(header http.Header) {
	headers := make([]string, 0, len(header))
	headers = append(headers, "host")
	for k, v := range header {
		canonicalKey := http.CanonicalHeaderKey(k)
		if shouldIgnore(canonicalKey, ignoredHeaders) {
			continue
		}
		if ctx.SignedHeaderVals == nil {
			ctx.SignedHeaderVals = make(http.Header, 3)
		}

		lowerCaseKey := strings.ToLower(k)
		if _, ok := ctx.SignedHeaderVals[lowerCaseKey]; ok {
			ctx.SignedHeaderVals[lowerCaseKey] = append(ctx.SignedHeaderVals[lowerCaseKey], v...)
			continue
		}

		headers = append(headers, lowerCaseKey)
		ctx.SignedHeaderVals[lowerCaseKey] = v
	}
	sort.Strings(headers)

	ctx.signedHeaders = strings.Join(headers, ";")

	headerValues := make([]string, len(headers))
	for i, k := range headers {
		if k == "host" {
			if ctx.Request.Host != "" {
				headerValues[i] = "host:" + ctx.Request.Host
			} else {
				headerValues[i] = "host:" + ctx.Request.URL.Host
			}
		} else {
			headerValues[i] = k + ":" +
				strings.Join(ctx.SignedHeaderVals[k], ",")
		}
	}
	stripExcessSpaces(headerValues)
	ctx.canonicalHeaders = strings.Join(headerValues, "\n")
}

func (ctx *signingCtx) buildCanonicalString() {
	uri := getURIPath(ctx.Request.URL)

	ctx.canonicalString = strings.Join([]string{
		ctx.Request.Method,
		uri,
		ctx.Request.URL.RawQuery,
		ctx.canonicalHeaders + "\n",
		ctx.signedHeaders,
		ctx.bodyDigest,
	}, "\n")
}

func (ctx *signingCtx) buildStringToSign() {
	ctx.stringToSign = strings.Join([]string{
		authHeaderPrefix,
		ctx.formattedTime,
		ctx.credentialString,
		hex.EncodeToString(makeSha256([]byte(ctx.canonicalString))),
	}, "\n")
}

func (ctx *signingCtx) buildSignature() {
	secret := ctx.credValues.Sk
	date := sha256Hmac([]byte("JDCLOUD2"+secret), []byte(ctx.formattedShortTime))
	region := sha256Hmac(date, []byte(ctx.Region))
	service := sha256Hmac(region, []byte(ctx.ServiceName))
	credentials := sha256Hmac(service, []byte("jdcloud2_request"))
	signature := sha256Hmac(credentials, []byte(ctx.stringToSign))
	ctx.signature = hex.EncodeToString(signature)
}

func (ctx *signingCtx) buildBodyDigest() {
	var hash string
	if ctx.Body == nil {
		hash = emptyStringSHA256
	} else {
		hash = hex.EncodeToString(makeSha256Reader(ctx.Body))
	}

	ctx.bodyDigest = hash
}

func (ctx *signingCtx) isRequestSigned() bool {
	if ctx.Request.Header.Get("Authorization") != "" {
		return true
	}

	return false
}

func sha256Hmac(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}

func makeSha256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func makeSha256Reader(reader io.ReadSeeker) []byte {
	hash := sha256.New()
	start, _ := reader.Seek(0, 1)
	defer reader.Seek(start, 0)

	io.Copy(hash, reader)
	return hash.Sum(nil)
}

const doubleSpace = "  "

func stripExcessSpaces(vals []string) {
	var j, k, l, m, spaces int
	for i, str := range vals {
		for j = len(str) - 1; j >= 0 && str[j] == ' '; j-- {
		}

		for k = 0; k < j && str[k] == ' '; k++ {
		}
		str = str[k : j+1]

		j = strings.Index(str, doubleSpace)
		if j < 0 {
			vals[i] = str
			continue
		}

		buf := []byte(str)
		for k, m, l = j, j, len(buf); k < l; k++ {
			if buf[k] == ' ' {
				if spaces == 0 {
					buf[m] = buf[k]
					m++
				}
				spaces++
			} else {
				spaces = 0
				buf[m] = buf[k]
				m++
			}
		}

		vals[i] = string(buf[:m])
	}
}

func getURIPath(u *url.URL) string {
	var uri string

	if len(u.Opaque) > 0 {
		uri = "/" + strings.Join(strings.Split(u.Opaque, "/")[3:], "/")
	} else {
		uri = u.EscapedPath()
	}

	if len(uri) == 0 {
		uri = "/"
	}

	return uri
}

func shouldIgnore(header string, ignoreHeaders []string) bool {
	for _, v := range ignoreHeaders {
		if v == header {
			return true
		}
	}
	return false
}
