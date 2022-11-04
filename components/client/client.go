package client

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grpc-boot/base"
)

type Client struct {
	conn          *websocket.Conn
	baseServerUri string
	level         uint8
	aes           *base.Aes
	key           []byte
	protocol      base.Protocol
}

func NewClient(uri string, level uint8, aes *base.Aes) (client *Client, err error) {
	client = &Client{
		baseServerUri: uri,
		level:         level,
		aes:           aes,
	}

	return client, err
}

func (c *Client) buildUri() (uri string, err error) {
	url := strings.Builder{}
	url.WriteString(c.baseServerUri)

	if strings.Index(c.baseServerUri, "?") > -1 {
		url.WriteByte('&')
	} else {
		url.WriteByte('?')
	}

	url.WriteString("l=")
	url.WriteString(strconv.Itoa(int(c.level)))

	switch c.level {
	case base.LevelV2:
		c.key = base.RandBytes(16)
		k := c.aes.CbcEncrypt(c.key)
		url.WriteString("&k=")
		url.WriteString(hex.EncodeToString(k))
	case base.LevelV1:
		key := make([]byte, 0, 32)
		key = append(key, base.RandBytes(16)...)
		key = append(key, base.RandBytes(16)...)
		k := c.aes.CbcEncrypt(key)

		if c.protocol == nil {
			c.protocol, err = base.NewV1(c.aes, k)
			if err != nil {
				return "", err
			}
		}

		url.WriteString("&k=")
		url.WriteString(hex.EncodeToString(k))
	default:
		if c.protocol == nil {
			c.protocol, err = base.NewV0()
			if err != nil {
				return "", err
			}
		}
	}

	return url.String(), nil
}

func (c *Client) Dial(timeout time.Duration) (err error) {
	serverUrl, err := c.buildUri()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ws, _, err := websocket.DefaultDialer.DialContext(ctx, serverUrl, nil)
	if err != nil {
		return err
	}

	c.conn = ws
	if c.level > base.LevelV1 {
		for {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				return err
			}

			pkg := &base.Package{}
			if err = pkg.Unpack(message); err != nil {
				return err
			}

			iv, err := base64.StdEncoding.DecodeString(pkg.Param.String("data"))
			if err != nil {
				return err
			}

			if pkg.Id != base.ConnectSuccess {
				continue
			}

			if c.protocol == nil {
				c.protocol, err = base.NewV2ForClient(c.aes, c.key, iv)
				if err != nil {
					return err
				}
			}

			break
		}
	}

	go c.watchMsg()

	return nil
}

func (c *Client) watchMsg() {
	defer c.conn.Close()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			base.Red("read msg with error:%s", err)
			return
		}

		base.Green("got msg: %s", message)
	}
}

func (c *Client) SendMsg(pkg *base.Package) error {
	return c.conn.WriteMessage(websocket.TextMessage, c.protocol.Pack(pkg))
}

func (c *Client) Close() error {
	return c.conn.Close()
}
