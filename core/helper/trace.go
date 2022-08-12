package helper

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	guest = "100"
)

var (
	paddingBytes = []byte("acegijlnprtuwybdfhkmoqsvxzegcgijlnprtuwacacijlnprtuwybdfhkmoqsvxzybdfhkmoqsvxz")
	maxLength    = len(paddingBytes)
)

func Id4Guest(clientIp uint32) string {
	return idWithIp(clientIp, guest)
}

func idWithIp(clientIp uint32, app string) string {
	var trace strings.Builder
	trace.WriteString(app)
	timeStr := strconv.FormatInt(time.Now().UnixNano(), 10)
	trace.WriteString(timeStr[:16])
	ipStr := strconv.FormatInt(int64(clientIp), 10)
	trace.WriteString(ipStr)

	paddingLength := 3
	if len(ipStr) < 10 {
		paddingLength = 13 - len(ipStr)
	}

	start := rand.Intn(maxLength - paddingLength)
	trace.Write(paddingBytes[start : start+paddingLength])

	return trace.String()
}
