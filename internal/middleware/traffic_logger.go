/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/base"
	"github.com/volcengine/key-proxy/internal/utils/logs"
	"time"
)

// StandardOnRequest is the standard onRequest hook.
func StandardOnRequest(ctx context.Context, requestInfo common.RequestInfo) {
	reqLog := base.DumpHttpRequest(requestInfo.Request)
	logs.CtxInfo(ctx, "[TrafficLogger] http request @%s, vendor: %s, cloud account id: %s, cloud account name: %s \n%s",
		requestInfo.RequestTime.String(),
		requestInfo.VendorName,
		requestInfo.CloudAccountId,
		requestInfo.CloudAccountName,
		reqLog,
	)
}

// StandardOnResponse is the standard onResponse hook.
func StandardOnResponse(ctx context.Context, response common.ResponseInfo) {
	logs.CtxInfo(ctx, "[TrafficLogger] http response @%s, cost: %dms, status: %d",
		response.ResponseTime.String(),
		response.Cost.Milliseconds(),
		response.HttpStatus,
	)
}

// TrafficLogger gets the data before and after the request is processed and pass the data into the hook functions.
func TrafficLogger(onRequest common.OnRequest, onResponse common.OnResponse) gin.HandlerFunc {
	return func(c *gin.Context) {
		mcdnArgs := base.GetMcdnArgs(c)
		commonInfo := base.GetBaseInfo(c)
		onRequest(c.Request.Context(), common.RequestInfo{
			BaseInfo:    commonInfo,
			RequestTime: mcdnArgs.RequestTime,
		})
		c.Header(base.ProxyVersionKey, base.Version)
		c.Next()
		// the response has been sent over here
		responseTime := time.Now()
		onResponse(c.Request.Context(), common.ResponseInfo{
			BaseInfo:               commonInfo,
			ResponseTime:           responseTime,
			Cost:                   responseTime.Sub(mcdnArgs.RequestTime),
			HttpStatus:             c.Writer.Status(),
			ProxyException:         c.GetBool(base.ProxyExceptionKey),
			ProxyExceptionTextCode: c.GetString(base.ProxyExceptionTextCodeKey),
		})
		return
	}
}
