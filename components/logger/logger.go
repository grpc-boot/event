package logger

import (
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/grpc-boot/base"
)

var (
	once           sync.Once
	defaultEncoder = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey: "Message",
		LevelKey:   "Level",
		TimeKey:    "DateTime",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		CallerKey:    "File",
		EncodeCaller: zapcore.ShortCallerEncoder,
	})
)

func InitLogger(infoSyncer zapcore.WriteSyncer, errorSyncer zapcore.WriteSyncer, encoder zapcore.Encoder, zapOpts ...zap.Option) {
	once.Do(func() {
		if encoder == nil {
			encoder = defaultEncoder
		}

		core := zapcore.NewTee(
			zapcore.NewCore(encoder, infoSyncer, zap.LevelEnablerFunc(func(z zapcore.Level) bool {
				return z >= zap.InfoLevel && z <= zap.WarnLevel
			})),
			zapcore.NewCore(encoder, errorSyncer, zap.LevelEnablerFunc(func(z zapcore.Level) bool {
				return z >= zap.ErrorLevel
			})),
		)
		base.InitZapWithCore(core, zapOpts...)
	})
}
