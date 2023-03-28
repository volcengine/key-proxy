/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package wangsu

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

func authorize(accountName string, passwd string) string {
	return base64.StdEncoding.EncodeToString([]byte(accountName + ":" + passwd))
}

func hmac64(sign string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha1.New, key)
	_, _ = h.Write([]byte(sign))
	sb := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(sb)
}
