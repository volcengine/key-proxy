/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package ucloud

import (
	"crypto/sha1"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

type Credential struct {
	Ak string
	Sk string
}

func (c *Credential) CreateSign(query string) string {
	payload := queryToMap(query)
	return sign(payload, c.Sk)
}

func (c *Credential) Apply(payload map[string]string) map[string]string {
	payload["PublicKey"] = c.Ak
	payload["Signature"] = sign(payload, c.Sk)
	return payload
}

func extractKeysSorted(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sign(params map[string]string, privateKey string) string {
	str := mapToStringSorted(params) + privateKey
	hashed := sha1.Sum([]byte(str))
	return hex.EncodeToString(hashed[:])
}

func mapToStringSorted(params map[string]string) string {
	sb := strings.Builder{}
	keys := extractKeysSorted(params)
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params[k])
	}
	return sb.String()
}

func queryToMap(query string) map[string]string {
	values := make(map[string]string, 4)
	urlValues, err := url.ParseQuery(query)
	if err != nil {
		return values
	}
	for k, v := range urlValues {
		if len(v) > 0 {
			values[k] = v[0]
		}
	}
	return values
}

func mapToQuery(values map[string]string) string {
	urlValues := make(url.Values, len(values))
	for k, v := range values {
		urlValues.Set(k, v)
	}
	return urlValues.Encode()
}
