package zapkey

import (
	"go.uber.org/zap"
)

func Event(name string) zap.Field {
	return zap.String("Event", name)
}

func Uri(uri string) zap.Field {
	return zap.String("Uri", uri)
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

func Address(addr string) zap.Field {
	return zap.String("Addr", addr)
}

func Value(value interface{}) zap.Field {
	return zap.Any("Value", value)
}

func Data(data []byte) zap.Field {
	return zap.ByteString("Data", data)
}
