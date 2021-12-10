package setecs

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

type SetEcs struct {
	ecsTablesLock   sync.RWMutex
	ecsBindingsLock sync.RWMutex
	Next            plugin.Handler
	debug           bool
	reload          time.Duration
	stopReload      chan struct{}
	ecsBindings     []*ecsBinding
	ecsTables       []*ecsTable
}

func NewSetEcs() *SetEcs {
	return &SetEcs{
		ecsTablesLock:   sync.RWMutex{},
		ecsBindingsLock: sync.RWMutex{},
		stopReload:      make(chan struct{}),
		ecsBindings:     make([]*ecsBinding, 0),
		ecsTables:       make([]*ecsTable, 0),
	}
}

func (se *SetEcs) MatchEcsTable(ipstr string) net.IP {
	se.ecsTablesLock.RLock()
	for _, table := range se.ecsTables {
		ecsip, ok := table.dict[ipstr]
		if ok {
			return ecsip
		}
	}
	se.ecsTablesLock.RUnlock()
	return nil
}

func (se *SetEcs) MatchEcsBinding(ip net.IP) net.IP {
	se.ecsBindingsLock.RLock()
	for _, bind := range se.ecsBindings {
		if bind.existIp(ip) {
			return bind.ecsip
		}
	}
	se.ecsBindingsLock.RUnlock()
	return nil
}

func (se *SetEcs) addEcsTable(t *ecsTable) {
	se.ecsTablesLock.Lock()
	se.ecsTables = append(se.ecsTables, t)
	se.ecsTablesLock.Unlock()
}

func (se *SetEcs) addEcsBinding(b *ecsBinding) {
	se.ecsBindingsLock.Lock()
	se.ecsBindings = append(se.ecsBindings, b)
	se.ecsBindingsLock.Unlock()
}

// ServeDNS implements the plugin.Handler interface.
func (se *SetEcs) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	clientIp := net.ParseIP(state.IP())
	if clientIp == nil {
		return plugin.NextOrFailure(state.Name(), se.Next, ctx, w, r)
	}

	var qHasECS = getMsgECS(r) != nil
	var wr = NewResponseReverter(w)
	var ecs *dns.EDNS0_SUBNET

	var ecsip = se.MatchEcsTable(state.IP())
	if ecsip != nil {
		ecsip = se.MatchEcsBinding(clientIp)
	}

	if ecsip != nil {
		if ip4 := ecsip.To4(); ip4 != nil { // is ipv4
			ecs = newEDNS0Subnet(ip4, 24, false)
		} else {
			if ip6 := ecsip.To16(); ip6 != nil { // is ipv6
				ecs = newEDNS0Subnet(ip6, 48, true)
			}
		}
	}

	// 强制设置 ECS， 如果请求本身没有 ECS， 那么响应中必须清除 ECS
	if ecs != nil {
		setECS(r, ecs)
		wr.removeEcs = qHasECS
	}

	return plugin.NextOrFailure(state.Name(), se.Next, ctx, wr, r)
}

func (se *SetEcs) Name() string { return "setecs" }

func (se *SetEcs) InlineEcsBinding() *ecsBinding {
	for _, binding := range se.ecsBindings {
		if binding.whichType == ItemTypeInline {
			return binding
		}
	}
	return nil
}

// 解析 ecsBindinbg
func (se *SetEcs) parseEcsBinding(ecsip string, items []string) error {
	ecsipb := net.ParseIP(ecsip)
	if ecsipb == nil {
		return fmt.Errorf("error ecsip %s", ecsip)
	}
	for _, item := range items {
		switch {
		case FileExists(item):
			eb := newEcsBinding(ItemTypePath, ecsipb, item, "")
			eb.loadFromFile()
			se.addEcsBinding(eb)
		case IsURL(item):
			eb := newEcsBinding(ItemTypeUrl, ecsipb, "", item)
			eb.loadFromUrl()
			se.addEcsBinding(eb)
		default:
			ipn, err := parseIpNet(item)
			if err != nil {
				log.Error(err)
				continue
			}
			eb := se.InlineEcsBinding()
			if eb == nil {
				eb = newEcsBinding(ItemTypeInline, ecsipb, "", "")
				se.addEcsBinding(eb)
			}
			eb.addInline(ipn)
		}
	}

	return nil
}

// 解析 ECS TABLE
func (se *SetEcs) parseEcsTable(items []string) error {
	for _, item := range items {
		switch {
		case FileExists(item):
			eb := newEcsTable(ItemTypePath, item, "")
			eb.loadFromFile()
			se.addEcsTable(eb)
		case IsURL(item):
			eb := newEcsTable(ItemTypePath, "", item)
			eb.loadFromFile()
			se.addEcsTable(eb)
		default:
			log.Errorf("ecs-table format error %s", item)
		}
	}
	return nil
}

func (se *SetEcs) periodicUpdate() {
	// Kick off initial name list content population
	if se.reload > 0 {
		go func() {
			ticker := time.NewTicker(se.reload)
			for {
				select {
				case <-se.stopReload:
					return
				case <-ticker.C:
					se.updateList()
				}
			}
		}()
	}
}

func (se *SetEcs) updateList() {

	for _, item := range se.ecsTables {
		switch item.whichType {
		case ItemTypePath:
			item.loadFromFile()
		case ItemTypeUrl:
			item.loadFromUrl()
		default:
		}
	}

	for _, item := range se.ecsBindings {
		switch item.whichType {
		case ItemTypePath:
			item.loadFromFile()
		case ItemTypeUrl:
			item.loadFromUrl()
		default:
		}
	}
}

func (se *SetEcs) OnStartup() error {
	se.periodicUpdate()
	return nil
}

func (se *SetEcs) OnShutdown() error {
	close(se.stopReload)
	return nil
}

func (se *SetEcs) debugPrint() {
	for _, s := range se.ecsBindings {
		log.Infof(s.String())
	}
	for _, s := range se.ecsTables {
		log.Infof(s.String())
	}
	log.Info("reload ", se.reload)
}
