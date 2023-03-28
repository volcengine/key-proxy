/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package handler

import (
	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	c.String(200, "Greeting")
}
