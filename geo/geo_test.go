package geo

import (
	"testing"
)

func TestIsIPInRadius(t *testing.T) {

	ip := "2001:0db8:1111:000a:00b0:0000:9000:0200"
	radius := 10000.0

	client := NewGeoClient("localhost", 6379, "", 0, "ipv4_location", "ipv6_location", "ipv6_location_info")

	res := client.IsIPInRadius(ip, "2001:0db8:0000:0000:abcd:0000:0000:1234", radius) // true
	t.Logf("target IP: %s, IP: %s, radius: %f, result %t", ip, "2001:0db8:0000:0000:abcd:0000:0000:1234", radius, res)
	res = client.IsIPInRadius(ip, "2001:0db8:cafe:0001:0000:0000:0000:0100", radius) // false
	t.Logf("target IP: %s, IP: %s, radius: %f, result %t", ip, "2001:0db8:cafe:0001:0000:0000:0000:0100", radius, res)
	res = client.IsIPInRadius(ip, "2001:0db8:cafe:0001:0000:0000:0000:0200", radius) // true
	t.Logf("target IP: %s, IP: %s, radius: %f, result %t", ip, "2001:0db8:cafe:0001:0000:0000:0000:0200", radius, res)
	res = client.IsIPInRadius(ip, "2001:0238:0000:0000:0000:0000:0000:0000", radius) // true
	t.Logf("target IP: %s, IP: %s, radius: %f, result %t", ip, "2001:0238:0000:0000:0000:0000:0000:0000", radius, res)

	radius = 50.0
	ip = "91.78.43.197"

	res = client.IsIPInRadius(ip, "91.78.144.24", radius) // true
	t.Logf("target IP: %s, IP: %s, radius: %f, result %t", ip, "91.78.144.24", radius, res)
	res = client.IsIPInRadius(ip, "91.78.40.24", radius) // false
	t.Logf("target IP: %s, IP: %s, radius: %f, result %t", ip, "91.78.40.24", radius, res)
	res = client.IsIPInRadius(ip, "91.78.80.18", radius) // true
	t.Logf("target IP: %s, IP: %s, radius: %f, result %t", ip, "91.78.80.18", radius, res)
	res = client.IsIPInRadius(ip, "91.78.224.24", radius) // true
	t.Logf("target IP: %s, IP: %s, radius: %f, result %t", ip, "91.78.224.24", radius, res)
	res = client.IsIPInRadius(ip, "91.75.168.21", radius) // false
	t.Logf("target IP: %s, IP: %s, radius: %f, result %t", ip, "91.75.168.21", radius, res)
}
