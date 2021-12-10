package setecs

import (
	"testing"

	"github.com/coredns/caddy"
)

func TestParse(t *testing.T) {
	c := caddy.NewTestController("dns", `setecs {
        ecs-binding 1.1.1.1 clients 127.0.0.1 172.21.66.0/24
        ecs-table ecs-tables.txt
        reload 10s
        debug
    }`)
	ecs, err := parseSetEcs(c)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ecs)
}
