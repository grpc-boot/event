package zapkey

import "go.uber.org/zap"

func Address(addr string) zap.Field {
	return zap.String("Address", addr)
}
