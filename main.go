package main

import (
	_ "github.com/ca17/edns/plugin/setecs"
	"github.com/coredns/coredns/core/dnsserver"
	_ "github.com/coredns/coredns/core/plugin"
	"github.com/coredns/coredns/coremain"
)

func index(slice []string, item string) int {
    for i := range slice {
        if slice[i] == item {
            return i
        }
    }
    return -1
}

func main() {
	idx := index(dnsserver.Directives, "forward")
	dnsserver.Directives = append(dnsserver.Directives[:idx], append([]string{"setecs"}, dnsserver.Directives[idx:]...)...)
	coremain.Run()
}
