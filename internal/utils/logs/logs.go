/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package logs

import (
	"context"
	"fmt"
	"github.com/volcengine/key-proxy/common"
	"github.com/volcengine/key-proxy/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
)

var _logger common.Logger

var logLevels = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
}

func MustInit(customLogger common.Logger) {
	if customLogger == nil {
		panic(fmt.Errorf("must init a logger"))
	}
	_logger = customLogger
}

func NewStandardLogger() (common.Logger, error) {
	logConf := config.Conf.Log
	filename, err := filepath.Abs(filepath.Join(logConf.Output, "./key_proxy.log"))
	if err != nil {
		return nil, err
	}
	maxAge := logConf.MaxAge
	if maxAge <= 0 {
		maxAge = 14
	}
	maxSize := logConf.MaxSize
	if maxSize <= 0 {
		maxSize = 100
	}
	writer := &lumberjack.Logger{
		Filename:  filename,
		MaxAge:    maxAge,
		MaxSize:   maxSize,
		LocalTime: false,
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	actualLevel, found := logLevels[logConf.Level]
	if !found {
		actualLevel = logLevels["info"]
	}
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), actualLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(writer), actualLevel),
	)
	return &StandardLogger{
		sugarLogger: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar(),
	}, nil
}

func CtxDebug(ctx context.Context, template string, args ...interface{}) {
	_logger.CtxDebug(ctx, template, args...)
}

func CtxInfo(ctx context.Context, template string, args ...interface{}) {
	_logger.CtxInfo(ctx, template, args...)
}

func CtxWarn(ctx context.Context, template string, args ...interface{}) {
	_logger.CtxWarn(ctx, template, args...)
}

func CtxError(ctx context.Context, template string, args ...interface{}) {
	_logger.CtxError(ctx, template, args...)
}

func CtxFatal(ctx context.Context, template string, args ...interface{}) {
	_logger.CtxFatal(ctx, template, args...)
}
