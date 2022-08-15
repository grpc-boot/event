package logger

import (
	"os"
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

func InitLoggerWithPath(level zapcore.Level, debugPath, infoPath, errorPath string, encoder zapcore.Encoder, zapOpts ...zap.Option) error {
	debugFile, err := os.OpenFile(debugPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	infoFile, err := os.OpenFile(infoPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	errorFile, err := os.OpenFile(errorPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	InitLogger(level, debugFile, infoFile, errorFile, encoder, zapOpts...)

	return nil
}

func InitLogger(level zapcore.Level, debugSyncer, infoSyncer, errorSyncer zapcore.WriteSyncer, encoder zapcore.Encoder, zapOpts ...zap.Option) {
	once.Do(func() {
		if encoder == nil {
			encoder = defaultEncoder
		}

		core := zapcore.NewTee(
			zapcore.NewCore(encoder, debugSyncer, zap.LevelEnablerFunc(func(z zapcore.Level) bool {
				return z >= level && z >= zap.DebugLevel && z < zap.InfoLevel
			})),
			zapcore.NewCore(encoder, infoSyncer, zap.LevelEnablerFunc(func(z zapcore.Level) bool {
				return z >= level && z >= zap.InfoLevel && z < zap.WarnLevel
			})),
			zapcore.NewCore(encoder, errorSyncer, zap.LevelEnablerFunc(func(z zapcore.Level) bool {
				return z >= level && z >= zap.WarnLevel
			})),
		)
		base.InitZapWithCore(core, zapOpts...)
	})
}
