/*
Copyright (2023) Beijing Volcano Engine Technology Ltd. All rights reserved.

Use of this source code is governed by the license that can be found in the LICENSE file.
*/

package logs

import (
	"context"
	"go.uber.org/zap"
)

// StandardLogger uses zap as the logger, implements the logger interface.
type StandardLogger struct {
	sugarLogger *zap.SugaredLogger
}

func (s *StandardLogger) Debug(template string, args ...interface{}) {
	s.sugarLogger.Debugf(template, args...)
}

func (s *StandardLogger) Info(template string, args ...interface{}) {
	s.sugarLogger.Infof(template, args...)
}

func (s *StandardLogger) Warn(template string, args ...interface{}) {
	s.sugarLogger.Warnf(template, args...)
}

func (s *StandardLogger) Error(template string, args ...interface{}) {
	s.sugarLogger.Errorf(template, args...)
}

func (s *StandardLogger) Fatal(template string, args ...interface{}) {
	s.sugarLogger.Fatalf(template, args...)
}

func (s *StandardLogger) CtxDebug(ctx context.Context, template string, args ...interface{}) {
	s.Debug(template, args...)
}
func (s *StandardLogger) CtxInfo(ctx context.Context, template string, args ...interface{}) {
	s.Info(template, args...)
}

func (s *StandardLogger) CtxWarn(ctx context.Context, template string, args ...interface{}) {
	s.Warn(template, args...)
}

func (s *StandardLogger) CtxError(ctx context.Context, template string, args ...interface{}) {
	s.Error(template, args...)
}

func (s *StandardLogger) CtxFatal(ctx context.Context, template string, args ...interface{}) {
	s.Fatal(template, args...)
}
