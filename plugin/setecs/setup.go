package setecs

import (
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

const PluginName = "setecs"

var log = clog.NewWithPlugin(PluginName)

func init() {
	plugin.Register(PluginName, setup)
}

func setup(c *caddy.Controller) error {
	p, err := parseSetEcs(c)
	if err != nil {
		return plugin.Error("setecs", err)
	}

	if p.debug {
		p.debugPrint()
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		p.Next = next
		return p
	})

	c.OnStartup(func() error {
		return p.OnStartup()
	})

	c.OnShutdown(func() error {
		return p.OnShutdown()
	})

	return nil
}

func parseSetEcs(c *caddy.Controller) (*SetEcs, error) {
	var secs = NewSetEcs()
	i := 0
	for c.Next() {
		if i > 0 {
			return nil, plugin.ErrOnce
		}
		i++
		for c.NextBlock() {
			switch c.Val() {
			case "ecs-binding":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen < 3 {
					return nil, c.Errf("format is `ecs-binding <1p> clients [ip(cidr) | filepath | url ...]`")
				}
				err := secs.parseEcsBinding(remaining[0], remaining[2:])
				if err != nil {
					return nil, c.Errf("parse client data error %s", err.Error())
				}
			case "ecs-table":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen < 1 {
					return nil, c.Errf("format is `ecs-table [ filepath | url ...]`")
				}
				err := secs.parseEcsTable(remaining)
				if err != nil {
					return nil, c.Errf("parse ecs-table data error %s", err.Error())
				}
			case "reload":
				remaining := c.RemainingArgs()
				if len(remaining) != 1 {
					return nil, c.Errf("reload needs a duration (zero seconds to disable)")
				}
				reload, err := time.ParseDuration(remaining[0])
				if err != nil {
					return secs, c.Errf("invalid duration for reload '%s'", remaining[0])
				}
				if reload < 0 {
					return nil, c.Errf("invalid negative duration for reload '%s'", remaining[0])
				}
				secs.reload = reload
			case "debug":
				secs.debug = true
			default:

			}
		}

	}
	return secs, nil
}
