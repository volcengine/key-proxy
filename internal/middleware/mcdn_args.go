/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/volcengine/key-proxy/internal/base"
)

// SetMcdnArgs get necessary parameters from the platform and then store them into gin context
func SetMcdnArgs() gin.HandlerFunc {
	return func(c *gin.Context) {
		mcdnArgs := base.NewMcdnArgs(c)
		c.Set(base.McdnArgsKey, mcdnArgs)
		c.Set(base.BaseInfoKey, base.NewBaseInfo(c, mcdnArgs))
		c.Next()
	}
}
