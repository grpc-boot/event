package helper

func Ip2Long(ipV4 string) uint32 {
	var (
		ip     uint32 = 0
		last          = -1
		offset        = 24
	)

	for start := 0; start < len(ipV4); start++ {
		if ipV4[start] >= '0' || ipV4[start] <= '9' {
			continue
		}

		if ipV4[start] != '.' || (start-1) == last {
			return 0
		}

		item := 0

	}

	ip = uint32(ips[0])<<24 + uint32(ips[1])<<16 + uint32(ips[2])<<8 + uint32(ips[3])

	return
}
