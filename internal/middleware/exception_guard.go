/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/base"
	"github.com/volcengine/key-proxy/internal/utils/logs"
	"time"
)

// ExceptionGuard captures all panics and recover while handling requests.
func ExceptionGuard(onResponse common.OnResponse) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if panicData := recover(); panicData != nil {
				logs.CtxError(c.Request.Context(), "capture an error: %v", panicData)
				mcdnArgs := base.GetMcdnArgs(c)
				var exception base.Exception
				if e, ok := panicData.(base.Exception); ok {
					exception = e
				} else {
					exception = base.InternalError.WithRawError(errors.New(fmt.Sprint(panicData)))
				}
				c.Set(base.ProxyExceptionTextCodeKey, exception.Code)
				c.Set(base.ProxyExceptionKey, true)
				errResponse := base.BuildErrorResponse(mcdnArgs, exception)
				if errResponse.ResponseMetadata.Error != nil {
					c.Header(base.ProxyStatusKey, base.ProxyStatusFailed)
				}
				responseTime := time.Now()
				baseInfo := base.GetBaseInfo(c)
				onResponse(c.Request.Context(), common.ResponseInfo{
					BaseInfo:               baseInfo,
					ResponseTime:           responseTime,
					Cost:                   responseTime.Sub(mcdnArgs.RequestTime),
					HttpStatus:             errResponse.ResponseMetadata.StatusCode,
					ProxyException:         c.GetBool(base.ProxyExceptionKey),
					ProxyExceptionTextCode: c.GetString(base.ProxyExceptionTextCodeKey),
				})
				c.JSON(errResponse.ResponseMetadata.StatusCode, errResponse)
			}
		}()
		c.Next()
	}
}
