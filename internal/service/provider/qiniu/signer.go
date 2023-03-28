/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package qiniu

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"github.com/volcengine/key-proxy/internal/utils"
	"net/http"
)

func getData(req *http.Request) ([]byte, error) {
	u := req.URL
	s := u.Path
	if u.RawQuery != "" {
		s += "?" + u.RawQuery
	}
	s += "\n"
	data := []byte(s)
	// only append data when content-type is form-encoded
	if req.Body != nil && req.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		b, err := utils.CopyRequestBody(req)
		if err != nil {
			return nil, err
		}
		data = append(data, b...)
	}
	return data, nil
}

func signToken(data []byte, ak, sk string) (token string) {
	h := hmac.New(sha1.New, []byte(sk))
	h.Write(data)
	sign := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return ak + ":" + sign
}

func getSignature(req *http.Request, ak, sk string) (token string, err error) {
	data, err := getData(req)
	if err != nil {
		return
	}
	token = signToken(data, ak, sk)
	return
}
