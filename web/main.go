package main

import (
	"fmt"
	"hash/crc32"
)

func main() {
	fmt.Println(crc32.ChecksumIEEE([]byte("helloasdf23@#Efddsfdsf23sDGsdg")))
	fmt.Println(crc32.ChecksumIEEE([]byte("worasdfLKNSDVLKNSDC*f3oadfasfld")))
}
