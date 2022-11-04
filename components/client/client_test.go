package client

import (
	"testing"
	"time"

	"github.com/grpc-boot/base"
)

var (
	aes        *base.Aes
	serverAddr = `ws://127.0.0.1:3333/ws`
)

func init() {
	aes, _ = base.NewAes("SD#$523asz7*&^df", "312c45cDvd$!F~12")
}

func TestClient_DialV0(t *testing.T) {
	client, err := NewClient(serverAddr, base.LevelJson, aes)
	if err != nil {
		t.Fatalf("want nil, got %s", err)
	}

	if err = client.Dial(time.Second); err != nil {
		t.Fatalf("want nil, got %s", err)
	}

	defer client.Close()

	tick := time.NewTicker(time.Second)

	num := 0
	for range tick.C {
		if num > 10 {
			break
		}

		err = client.SendMsg(&base.Package{
			Id:   base.Login,
			Name: "login",
			Param: base.JsonParam{
				"token": time.Now().String(),
			},
		})

		if err != nil {
			t.Fatalf("want nil, got %s", err)
		}
		num++
	}
}

func TestClient_DialV1(t *testing.T) {
	client, err := NewClient(serverAddr, base.LevelV1, aes)
	if err != nil {
		t.Fatalf("want nil, got %s", err)
	}

	if err = client.Dial(time.Second); err != nil {
		t.Fatalf("want nil, got %s", err)
	}

	defer client.Close()

	tick := time.NewTicker(time.Second)

	num := 0
	for range tick.C {
		if num > 10 {
			break
		}

		err = client.SendMsg(&base.Package{
			Id:   base.Login,
			Name: "login",
			Param: base.JsonParam{
				"token": time.Now().String(),
			},
		})

		if err != nil {
			t.Fatalf("want nil, got %s", err)
		}
		num++
	}
}

func TestClient_DialV2(t *testing.T) {
	client, err := NewClient(serverAddr, base.LevelV2, aes)
	if err != nil {
		t.Fatalf("want nil, got %s", err)
	}

	if err = client.Dial(time.Second); err != nil {
		t.Fatalf("want nil, got %s", err)
	}

	defer client.Close()

	tick := time.NewTicker(time.Second)

	num := 0
	for range tick.C {
		if num > 10 {
			break
		}

		err = client.SendMsg(&base.Package{
			Id:   base.Login,
			Name: "login",
			Param: base.JsonParam{
				"token": time.Now().String(),
			},
		})

		if err != nil {
			t.Fatalf("want nil, got %s", err)
		}
		num++
	}
}
