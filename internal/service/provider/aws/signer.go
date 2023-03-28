/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package aws

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/volcengine/key-proxy/internal/utils"
	"net/http"
	"time"
)

func signRequest(ctx context.Context, req *http.Request, ak, sk, accessToken, region string, service string, signTime time.Time) error {
	body, err := utils.CopyRequestBody(req)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(body)
	seeker := aws.ReadSeekCloser(reader)
	cre := credentials.NewStaticCredentials(ak, sk, accessToken)
	signer := v4.NewSigner(cre)
	_, err = signer.Sign(req, seeker, service, region, signTime)
	if err != nil {
		return err
	}
	return err
}
