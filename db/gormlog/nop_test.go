package gormlog_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	gormlogger "gorm.io/gorm/logger"

	"github.com/brpaz/lib-go/db/gormlog"
)

func TestNewNopLogger(t *testing.T) {
	t.Parallel()

	l := gormlog.NewNopLogger()

	// Must not panic and must implement gormlogger.Interface.
	var iface gormlogger.Interface = l
	iface.Info(context.Background(), "msg")
	iface.Warn(context.Background(), "msg")
	iface.Error(context.Background(), "msg")
	iface.Trace(context.Background(), time.Now(), func() (string, int64) { return "", 0 }, nil)

	cloned := iface.LogMode(gormlogger.Silent)
	assert.NotNil(t, cloned)
}
