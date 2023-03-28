/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package base

import (
	"fmt"
)

type Exception struct {
	error
	StatusCode int
	Code       string
	Message    string
	MessageCn  string
	RawError   string // 原始错误信息字符串
}

const textCodePrefix = "Proxy."

func NewException(statusCode int, textCode, msg string, msgCn string) Exception {
	return Exception{
		StatusCode: statusCode,
		Code:       textCodePrefix + textCode,
		Message:    msg,
		MessageCn:  msgCn,
		RawError:   "",
	}
}

func (e Exception) Error() string {
	return fmt.Sprintf("StatusCode: %v, Code: %v, Message: %v, RawError: %v", e.StatusCode, e.Code, e.Message, e.RawError)
}

func (e Exception) Unwrap() error {
	return e.error
}

func (e Exception) Is(target Exception) bool {
	return e.Code == target.Code
}

func (e Exception) WithRawError(err error) Exception {
	e.error = err
	e.RawError = ""
	if e.error != nil {
		e.RawError = e.error.Error()
	}
	return e
}

func (e Exception) WithRawErrorStr(s string) Exception {
	e.RawError = s
	return e
}

func (e Exception) WithStatusCode(code int) Exception {
	e.StatusCode = code
	return e
}

func (e Exception) WithCode(code string) Exception {
	e.Code = code
	return e
}

func (e Exception) WithMessage(msg string) Exception {
	e.Message = msg
	return e
}

func (e Exception) WithMessageCn(msg string) Exception {
	e.MessageCn = msg
	return e
}
