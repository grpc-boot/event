package helper

import (
	"testing"

	"github.com/grpc-boot/base"
)

func TestIdWithIp(t *testing.T) {
	longIp, _ := base.Ip2Long("127.0.0.1")

	t.Logf("ip value:%d\n", longIp)
	id := Id4Guest(longIp)
	t.Logf("%s %d\n", id, len(id))

	longIp, _ = base.Ip2Long("145.234.23.2")

	t.Logf("ip value:%d\n", longIp)
	id = Id4Guest(longIp)
	t.Logf("%s %d\n", id, len(id))
}
