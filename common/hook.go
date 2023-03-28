package common

import (
	"context"
	"net/http"
	"time"
)

type BaseInfo struct {
	Request          *http.Request
	RequestTime      time.Time
	CloudAccountId   string
	CloudAccountName string
	SubProduct       string
	TopAccountId     string
	VendorName       string
	RequestId        string
	TargetUrl        string
	ProxyVersion     string
}

type OnRequest func(ctx context.Context, requestInfo RequestInfo)
type OnResponse func(ctx context.Context, response ResponseInfo)

type RequestInfo struct {
	BaseInfo
	RequestTime time.Time
}

type ResponseInfo struct {
	BaseInfo
	ResponseTime           time.Time
	Cost                   time.Duration
	HttpStatus             int
	ProxyException         bool
	ProxyExceptionTextCode string
}
