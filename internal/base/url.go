/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package base

import (
	"net/url"
	"strings"
)

// QuickEncode encodes queries without sorting
func QuickEncode(q url.Values) string {
	if q == nil {
		return ""
	}
	var buf strings.Builder
	for k := range q {
		vs := q[k]
		keyEscaped := url.QueryEscape(k)
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(v))
		}
	}
	return buf.String()
}
