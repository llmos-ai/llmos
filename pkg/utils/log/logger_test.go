package log_test

import (
	"context"
	"log/slog"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/llmos-ai/llmos/pkg/utils/log"
)

var _ = Describe("logger", Label("log", "logger"), func() {
	l1 := log.NewLogger(context.Background())
	It("TestNewLogger returns a logger interface", func() {
		l2 := slog.Default()
		Expect(reflect.TypeOf(l1.GetLogger()).Kind()).To(Equal(reflect.TypeOf(l2).Kind()))
	})
})
