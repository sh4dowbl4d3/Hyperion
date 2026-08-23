package ssrf

import "net"

func netParseIP(s string) net.IP { return net.ParseIP(s) }
