package utils

import (
	"context"
	"github.com/CranePeng/fenv-middleware/utils/common"
	"github.com/CranePeng/fenv-middleware/utils/logger"
	"testing"
)

func TestLogDemo(t *testing.T) {
	conf := logger.Config{Ctx: context.TODO(), LogLevel: logger.Info, CreateFile: true}
	log := logger.New(conf)
	log.Info(conf.Ctx, "测试 %v", "asf")
	log.Debug(conf.Ctx, "测试")
	log.Warn(conf.Ctx, "测试")
	common.GetCurrentAbPath()
}
