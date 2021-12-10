package setecs

import (
	"testing"
)

func Test_parseIpNet(t *testing.T) {
	ip , err := parseIpNet("192.168.0.1/32")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip.String())
	t.Log(ip.FirstAddress())
	t.Log(ip.LastAddress())
}
