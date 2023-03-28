/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package base

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/utils"
	"net/http"
	"time"
)

const (
	ProxyStatusKey            = "X-Mcdn-Proxy-Status"
	ProxyVersionKey           = "X-Mcdn-Proxy-Version"
	RequestIdKey              = "X-Mcdn-Request-Id"
	CloudAccountIdKey         = "X-Mcdn-Cloud-Account-Id"
	OriginHostKey             = "X-Origin-Host"
	CloudAccountNameKey       = "X-Mcdn-Cloud-Account-Name"
	TopAccountIdKey           = "X-Mcdn-Top-Account-Id"
	OriginUrlKey              = "X-Mcdn-Origin-Uri"
	KeptHeaders               = "X-Mcdn-Kept-Headers"
	VendorNameKey             = "X-Mcdn-Vendor-Name"
	SubProductKey             = "X-Mcdn-Cloud-Account-SubProject"
	McdnArgsKey               = "McdnArgs"
	BaseInfoKey               = "BaseInfo"
	ProxyStatusFailed         = "Failed"
	ProxyExceptionTextCodeKey = "X-Exception-TextCode"
	ProxyExceptionKey         = "X-Proxy-Exception"
)

type ProxyResponse struct {
	ResponseMetadata ResponseMetadata `json:",omitempty"`
}

type ResponseMetadata struct {
	Error      *ErrorObj `json:",omitempty"`
	RequestId  string    `json:",omitempty"`
	Version    string    `json:",omitempty"`
	StatusCode int       `json:"-"`
}

type ErrorObj struct {
	Code    string
	Message string
	Detail  string `json:",omitempty"`
}

type TraceInfo struct {
	RequestId   string
	RequestTime time.Time
}

type McdnArgs struct {
	TraceInfo
	CloudAccountId   string
	TopAccountId     string
	VendorName       string
	CloudAccountName string
	SubProduct       string
	Version          string
}

func (e *ErrorObj) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("Code: %v, Message: %v", e.Code, e.Message)
}

func NewMcdnArgs(c *gin.Context) McdnArgs {
	return McdnArgs{
		TraceInfo: TraceInfo{
			RequestId:   c.GetHeader(RequestIdKey),
			RequestTime: time.Now(),
		},
		CloudAccountId:   c.GetHeader(CloudAccountIdKey),
		TopAccountId:     c.GetHeader(TopAccountIdKey),
		VendorName:       c.GetHeader(VendorNameKey),
		CloudAccountName: c.GetHeader(CloudAccountNameKey),
		SubProduct:       c.GetHeader(SubProductKey),
		Version:          Version,
	}
}

func NewBaseInfo(c *gin.Context, mcdnArgs McdnArgs) common.BaseInfo {
	return common.BaseInfo{
		Request:          c.Request,
		RequestTime:      mcdnArgs.RequestTime,
		CloudAccountId:   mcdnArgs.CloudAccountId,
		CloudAccountName: mcdnArgs.CloudAccountName,
		SubProduct:       mcdnArgs.SubProduct,
		TopAccountId:     mcdnArgs.TopAccountId,
		VendorName:       mcdnArgs.VendorName,
		RequestId:        mcdnArgs.RequestId,
		TargetUrl:        c.GetHeader(OriginUrlKey),
		ProxyVersion:     Version,
	}
}

func GetMcdnArgs(c *gin.Context) McdnArgs {
	val, existed := c.Get(McdnArgsKey)
	if !existed {
		return NewMcdnArgs(c)
	}
	return val.(McdnArgs)
}

func GetBaseInfo(c *gin.Context) common.BaseInfo {
	val, existed := c.Get(BaseInfoKey)
	if !existed {
		return NewBaseInfo(c, GetMcdnArgs(c))
	}
	return val.(common.BaseInfo)
}

func BuildErrorResponse(args McdnArgs, except Exception) *ProxyResponse {
	return &ProxyResponse{
		ResponseMetadata: ResponseMetadata{
			Error: &ErrorObj{
				Code:    except.Code,
				Message: except.Message,
				Detail:  except.RawError,
			},
			RequestId:  args.RequestId,
			Version:    args.Version,
			StatusCode: except.StatusCode,
		},
	}
}

func DumpHttpRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	if cmd, err := utils.GetCurlCommand(req); err == nil {
		return cmd.String()
	}
	return ""
}
