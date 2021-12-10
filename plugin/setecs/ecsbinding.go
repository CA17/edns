package setecs

import (
	"bufio"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/c-robinson/iplib"
)

const (
	ItemTypePath = iota
	ItemTypeUrl
	ItemTypeInline // Dummy
)

type ecsBinding struct {
	sync.RWMutex
	whichType   int
	path        string
	mtime       time.Time
	size        int64
	url         string
	contentHash uint64
	ecsip       net.IP
	clients     []iplib.Net
	inline      []iplib.Net
}

func newEcsBinding(wtype int, ecsip net.IP, path, url string) *ecsBinding {
	return &ecsBinding{
		RWMutex:     sync.RWMutex{},
		whichType:   wtype,
		path:        path,
		mtime:       time.Time{},
		size:        0,
		url:         url,
		contentHash: 0,
		ecsip:       ecsip,
		clients:     make([]iplib.Net, 0),
		inline:      make([]iplib.Net, 0),
	}
}

func (eb *ecsBinding) existIpNet(inet iplib.Net) bool {
	eb.RLock()
	defer eb.RUnlock()
	for _, ipnet := range eb.clients {
		if ipnet.ContainsNet(inet) {
			return true
		}
	}
	return false
}

func (eb *ecsBinding) existIp(ip net.IP) bool {
	eb.RLock()
	defer eb.RUnlock()
	for _, ipnet := range eb.clients {
		if ipnet.Contains(ip) {
			return true
		}
	}
	return false
}

func (eb *ecsBinding) addInline(client iplib.Net) bool {
	if eb.existIpNet(client) {
		return false
	}
	eb.Lock()
	defer eb.Unlock()
	eb.inline = append(eb.inline, client)
	return true
}

func (eb *ecsBinding) parentIpNet(inet iplib.Net) iplib.Net {
	eb.RLock()
	defer eb.RUnlock()
	for _, ipnet := range eb.clients {
		if ipnet.ContainsNet(inet) {
			return ipnet
		}
	}
	return nil
}

func (eb *ecsBinding) loadFromFile() {
	if eb.whichType != ItemTypePath || len(eb.path) == 0 {
		return
	}
	file, err := os.Open(eb.path)
	if err != nil {
		log.Errorf("clients file read error %s", eb.path)
		return
	}

	stat, err := file.Stat()
	if err == nil {
		eb.RLock()
		mtime := eb.mtime
		size := eb.size
		eb.RUnlock()

		if stat.ModTime() == mtime && stat.Size() == size {
			return
		}
	} else {
		// Proceed parsing anyway
		log.Warningf("%v", err)
	}

	t1 := time.Now()
	addrs, totalLines := eb.parse(file)
	t2 := time.Since(t1)
	log.Debugf("Parsed %v  time spent: %v name added: %v / %v", file.Name(), t2, len(addrs), totalLines)

	eb.Lock()
	eb.clients = addrs
	eb.mtime = stat.ModTime()
	eb.size = stat.Size()
	eb.Unlock()
}

func (eb *ecsBinding) parse(r io.Reader) ([]iplib.Net, uint64) {
	addrs := make([]iplib.Net, 0)
	var totalLines uint64
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		totalLines++

		line := scanner.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}

		addr, err := ParseIpNet(line)
		if err != nil {
			log.Errorf("error netaddr %s %s", line, err.Error())
			continue
		}

		for _, inet := range addrs {
			if addr.ContainsNet(inet) {
				log.Warningf("%s ContainsNet %s", addr.String(), inet.String())
				continue
			}
		}
		addrs = append(addrs, addr)
	}

	return addrs, totalLines
}

func (eb *ecsBinding) loadFromUrl() {
	if eb.whichType != ItemTypeUrl || len(eb.url) == 0 {
		return
	}

	t1 := time.Now()
	content, err := HttpGet(eb.url, nil)
	t2 := time.Since(t1)
	if err != nil {
		log.Warningf("Failed to update %q, err: %v", eb.url, err)
		return
	}
	contentStr := string(content)

	eb.RLock()
	contentHash := eb.contentHash
	eb.RUnlock()
	contentHash1 := StringHash(contentStr)
	if contentHash1 == contentHash {
		return
	}

	addrs := make([]iplib.Net, 0)
	var totalLines uint64
	t3 := time.Now()
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		totalLines++
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		addr, err := ParseIpNet(line)
		if err != nil {
			log.Errorf("error netaddr %s %s", line, err.Error())
			continue
		}
		for _, inet := range addrs {
			if addr.ContainsNet(inet) {
				log.Warningf("%s ContainsNet %s", addr.String(), inet.String())
				continue
			}
		}
		addrs = append(addrs, addr)
	}
	t4 := time.Since(t3)
	log.Debugf("Fetched %v, time spent: %v %v, added: %v / %v, hash: %#x",
		eb.url, t2, t4, len(addrs), totalLines, contentHash1)

	eb.Lock()
	eb.clients = addrs
	eb.contentHash = contentHash1
	eb.Unlock()
}

func (eb *ecsBinding) String() string {
	sb := strings.Builder{}
	sb.WriteString("ecsBinding:ecsip=")
	sb.WriteString(eb.ecsip.String())
	sb.WriteString(";clients=")
	c := 0
	for _, client := range eb.clients {
		if c >= 5 {
			sb.WriteString("......")
			break
		}
		sb.WriteString(client.String())
		sb.WriteString(",")
		c += 1
	}
	sb.WriteString(";inlines=")
	cc := 0
	for _, il := range eb.inline {
		if cc >= 5 {
			sb.WriteString("......")
			break
		}
		sb.WriteString(il.String())
		sb.WriteString(",")
		cc += 1
	}
	return sb.String()
}
