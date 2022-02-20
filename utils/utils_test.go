package utils

import (
	"context"
	"github.com/CranePeng/fenv-middleware/utils/common"
	"github.com/CranePeng/fenv-middleware/utils/logger"
	"path"
	"path/filepath"
	"testing"
)

func TestLogDemo(t *testing.T) {
	x := common.GetCurrentAbPath()
	x, _ = path.Split(x)
	x = filepath.Dir(x)
	x = filepath.Dir(x)
	x = path.Join(x, "logs")
	conf := logger.Config{Ctx: context.TODO(), LogLevel: logger.Debug, CreateFile: true, LogPath: x}
	log := logger.New(conf)
	log.Info(conf.Ctx, "测试 %v", "asf")
	log.Debug(conf.Ctx, "测试")
	log.Warn(conf.Ctx, "测试")
	//log.SetFormat(&nested.Formatter{
	//	TimestampFormat: time.RFC3339,
	//	FieldsOrder:     []string{"name", "age"},
	//})
	//log.WithField(logrus.Fields{
	//	"name": "dj",
	//	"age":  18,
	//})
	//log.Info(conf.Ctx,"测试2")
}
