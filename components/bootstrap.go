package components

import (
	"math/rand"
	"time"

	"event/lib/constant"

	"github.com/grpc-boot/base"
)

var (
	conf *base.Config
)

func Bootstrap() {
	rand.Seed(time.Now().UnixNano())

	conf = base.DefaultContainer.Config()

	loadAes()
	loadAccept()
}

func loadAes() {
	aesKey := conf.Params.String("aes.key")
	if len(aesKey) != 32 {
		base.RedFatal("aes key length is not 32")
	}
	aes, err := base.NewAes(aesKey[0:16], aesKey[16:])
	if err != nil {
		base.RedFatal("load aes failed")
	}
	base.DefaultContainer.Set(constant.Aes, aes)
}

func loadAccept() {
	aes, exists := base.DefaultContainer.Get(constant.Aes)
	if !exists {
		base.RedFatal("aes not exists")
	}

	level := uint8(conf.Params.Int64("accept.level"))
	accept := base.NewAccept(aes.(*base.Aes), level)
	base.DefaultContainer.Set(constant.Accept, accept)
}
