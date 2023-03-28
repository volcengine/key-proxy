/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package base

var (
	InternalError                 = NewException(500, "InternalError", "There was an internal error occurred.", "内部错误，请重试或联系客服人员解决。")
	CloudAccountNotFound          = NewException(404, "CloudAccountNotFound", "The cloud account is not found in the config.", "在配置中未找到对应的账号。")
	ValidateCredentialErr         = NewException(401, "ValidateCredentialErr", "The proxy credential provided does not match the configuration.", "提供的代理秘钥与配置不符。")
	ResignInternalErr             = NewException(500, "ResignInternalErr", "There was an internal error occurred during resigning.", "签算时发生内部错误。")
	ValidateCredentialInternalErr = NewException(500, "ValidateCredentialInternalErr", "There was an internal error occurred during validating.", "验证签算时发生内部错误。")
	ReformRequestInternalErr      = NewException(500, "ReformRequestInternalErr", "There was an internal error occurred during reforming the request.", "处理请求时发生内部错误。")
	NetworkErr                    = NewException(502, "NetworkErr", "There was a network error occurred during requesting.", "请求厂商时发生网络错误。")
)
