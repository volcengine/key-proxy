/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"net/http"
	"net/url"
	"strings"

	"github.com/volcengine/key-proxy/internal/base"
	"github.com/volcengine/key-proxy/internal/config"
	"github.com/volcengine/key-proxy/internal/utils/logs"
)

var providers = make(map[string]RegisterFunc, 12)

type RegisterFunc func(credential common.Credentials) IProvider

func RegisterProvider(vendor string, f RegisterFunc) {
	providers[vendor] = f
}

type IProviderService interface {
	ReformRequest(ctx context.Context, req *http.Request)
}

type IProvider interface {
	ValidateRequest(ctx context.Context, req *http.Request) (context.Context, bool, error)
	ResignRequest(ctx context.Context, req *http.Request) error
	String() string
}

type ImplProviderService struct {
	endpointProviders map[string]IProvider
}

// New registers cloud vendor providers to the service.
func New(endpoints []common.Endpoint) (*ImplProviderService, error) {
	s := &ImplProviderService{
		endpointProviders: make(map[string]IProvider, 10),
	}
	for _, endpoint := range endpoints {
		if endpoint.CloudAccountName == "" {
			return nil, errors.New("the name of cloud account cannot be empty")
		}
		registerFunc, found := providers[endpoint.Vendor]
		if !found {
			availableVendorCodes := make([]string, 0, len(providers))
			for vendorCode := range providers {
				availableVendorCodes = append(availableVendorCodes, string(vendorCode))
			}
			return nil, fmt.Errorf("unknown vendor code: \"%s\", available vendor codes are: [%s]", endpoint.Vendor, strings.Join(availableVendorCodes, ", "))
		}
		provider := registerFunc(endpoint.Credentials)
		// duplicated cloud account name is forbidden
		_, existed := s.endpointProviders[endpoint.CloudAccountName]
		if existed {
			return nil, fmt.Errorf("cloud account has existed, duplicated name: %s", endpoint.CloudAccountName)
		}
		s.endpointProviders[endpoint.CloudAccountName] = provider
		logs.CtxInfo(context.Background(), "loaded %s provider with cloud account (name: %v) successfully", endpoint.Vendor, endpoint.CloudAccountName)
	}

	return s, nil
}

// reformRequest reverts request to the normal request to cloud vendors.
func (s *ImplProviderService) reformRequest(req *http.Request) error {
	originUri := req.Header.Get(base.OriginUrlKey)
	keptHeaders := req.Header.Get(base.KeptHeaders)
	headers := strings.Split(strings.ToLower(keptHeaders), ",")
	for _, keptHeader := range headers {
		if keptHeader == "host" {
			req.Header.Set("Host", req.Header.Get(base.OriginHostKey))
		}
	}
	for key := range req.Header {
		keep := false
		for _, keptKey := range headers {
			if strings.ToLower(key) == keptKey {
				keep = true
				break
			}
		}
		if !keep {
			req.Header.Del(key)
		}
	}
	req.Header.Del(base.OriginHostKey)
	req.Header.Del(base.CloudAccountIdKey)
	req.Header.Del(base.CloudAccountNameKey)
	req.Header.Del(base.KeptHeaders)
	req.Header.Del(base.OriginUrlKey)
	req.Header.Del(base.VendorNameKey)
	req.Header.Del(base.TopAccountIdKey)
	req.Header.Del(base.RequestIdKey)

	u, err := url.Parse(originUri)
	if err != nil {
		return err
	}
	req.URL = u
	req.Host = u.Host
	return nil
}

// ReformRequest reverts the format of the request from the platform, then verify and resign it.
func (s *ImplProviderService) ReformRequest(ctx context.Context, req *http.Request) {
	cloudAccountName := req.Header.Get(base.CloudAccountNameKey)
	provider, found := s.getEndpointProvider(cloudAccountName)

	forbidden := config.Conf.Forbidden
	if !found && forbidden.ForbiddenAccountNotFound {
		panic(base.CloudAccountNotFound.WithRawError(fmt.Errorf("cloud account is not found, name: %s", cloudAccountName)))
	}

	err := s.reformRequest(req)
	if err != nil {
		panic(base.ReformRequestInternalErr.WithRawError(err))
	}

	// skip validation and resign, forward the request directly if provider was not found
	if provider == nil {
		return
	}
	ctx, ok, err := provider.ValidateRequest(ctx, req)
	if err != nil {
		panic(base.ValidateCredentialInternalErr.WithRawError(err))
	}
	if !ok && forbidden.ForbiddenProxyCredentialErr {
		panic(base.ValidateCredentialErr.WithRawError(fmt.Errorf("[%s] proxy ak or sk is wrong", provider.String())))
	}
	if ok {
		// resign the request, if this request was valid
		err = provider.ResignRequest(ctx, req)
		if err != nil {
			panic(base.ResignInternalErr.WithRawError(err))
		}
	}
}

func (s *ImplProviderService) getEndpointProvider(cloudAccountName string) (IProvider, bool) {
	provider, ok := s.endpointProviders[cloudAccountName]
	if !ok {
		return nil, false
	}
	return provider, true
}
