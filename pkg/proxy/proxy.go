/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package proxy

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/base"
	"github.com/volcengine/key-proxy/internal/config"
	"github.com/volcengine/key-proxy/internal/handler"
	"github.com/volcengine/key-proxy/internal/middleware"
	"github.com/volcengine/key-proxy/internal/service"
	"github.com/volcengine/key-proxy/internal/utils/logs"
	"net/http"
	"net/http/httputil"
)

var (
	defaultTransport = http.DefaultTransport.(*http.Transport).Clone()
)

type Option struct {
	Config                *common.Config
	Logger                common.Logger
	OnRequestHook         common.OnRequest
	OnReformedRequestHook common.OnRequest
	OnResponseHook        common.OnResponse
}

type withOption func(o *Option)

func WithLogger(logger common.Logger) withOption {
	return func(o *Option) {
		o.Logger = logger
	}
}

func WithOnRequestHook(hook common.OnRequest) withOption {
	return func(o *Option) {
		o.OnRequestHook = hook
	}
}

func WithOnResponseHook(hook common.OnResponse) withOption {
	return func(o *Option) {
		o.OnResponseHook = hook
	}
}

type KeyProxy struct {
	opt Option
}

func New(conf *common.Config, opts ...withOption) (*KeyProxy, error) {
	option := Option{Config: conf}
	for _, opt := range opts {
		opt(&option)
	}
	if option.Logger == nil {
		standardLogger, err := logs.NewStandardLogger()
		if err != nil {
			return nil, err
		}
		option.Logger = standardLogger
	}
	if option.OnRequestHook == nil {
		option.OnRequestHook = middleware.StandardOnRequest
	}
	if option.OnResponseHook == nil {
		option.OnResponseHook = middleware.StandardOnResponse
	}
	logs.MustInit(option.Logger)
	s := &KeyProxy{
		opt: option,
	}
	err := s.Reload(s.opt.Config)
	return s, err
}

func (s *KeyProxy) Run() error {
	return s.RegisterHttp()
}

func (s *KeyProxy) Reload(conf *common.Config) error {
	err := service.Init(conf)
	if err != nil {
		return err
	}
	config.MustInit(conf)
	return nil
}

func (s *KeyProxy) RegisterHttp() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.SetMcdnArgs())
	r.Use(middleware.TrafficLogger(s.opt.OnRequestHook, s.opt.OnResponseHook))
	r.Use(middleware.ExceptionGuard(s.opt.OnResponseHook))
	s.customizeRegister(r)

	var err error
	httpConf := s.opt.Config.Http
	tlsConf := httpConf.Tls

	ctx := context.Background()
	if tlsConf.Enabled {
		logs.CtxInfo(ctx, "launch https server on %v", tlsConf.Address)
		if err = r.RunTLS(tlsConf.Address, tlsConf.CertFile, tlsConf.KeyFile); err != nil {
			return err
		}
	} else {
		logs.CtxInfo(ctx, "launch http server on %v", httpConf.Address)
		if err = r.Run(httpConf.Address); err != nil {
			return err
		}
	}
	return nil
}

func (s *KeyProxy) customizeRegister(r *gin.Engine) {
	r.GET("/ping", handler.Ping)
	{
		p := new(httputil.ReverseProxy)
		defaultTransport.MaxIdleConns = 200
		defaultTransport.MaxConnsPerHost = 100
		defaultTransport.MaxIdleConnsPerHost = 100
		p.Transport = defaultTransport
		p.Director = func(req *http.Request) {
			providerService := service.GetProviderService()
			providerService.ReformRequest(req.Context(), req)
			logs.CtxInfo(req.Context(), "reformed request: %s", base.DumpHttpRequest(req))
		}
		p.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
			if err != nil {
				panic(base.NetworkErr.WithRawError(err))
			}
		}
		r.NoRoute(func(c *gin.Context) {
			p.ServeHTTP(c.Writer, c.Request)
		})
	}
}
