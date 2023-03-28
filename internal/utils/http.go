/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package utils

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

type CurlCommand struct {
	slice []string
}

func (c *CurlCommand) append(newSlice ...string) {
	c.slice = append(c.slice, newSlice...)
}

func (c *CurlCommand) String() string {
	return strings.Join(c.slice, " ")
}

func bashEscape(str string) string {
	return `'` + strings.Replace(str, `'`, `'\''`, -1) + `'`
}

func GetCurlCommand(req *http.Request) (*CurlCommand, error) {
	command := CurlCommand{
		slice: make([]string, 0, 10),
	}

	command.append("curl")

	command.append("-X", bashEscape(req.Method))

	if req.Body != nil {
		body, err := CopyRequestBody(req)
		if err != nil {
			return nil, err
		}
		bodyEscaped := bashEscape(string(body))
		command.append("-d", bodyEscaped)
	}

	keys := make([]string, 0, len(req.Header))

	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := strings.Join(req.Header[k], " ")
		if strings.ToLower(k) == "authorization" {
			tmp, _ := base64.StdEncoding.DecodeString(v)
			v = strings.Split(string(tmp), ":")[0]
		}
		command.append("-H", bashEscape(k+": "+v))
	}

	command.append(bashEscape(req.URL.String()))

	return &command, nil
}

func CopyRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return []byte{}, nil
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	err = r.Body.Close()
	if err != nil {
		return nil, err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	return data, nil
}
