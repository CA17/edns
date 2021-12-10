package setecs

import (
	"bufio"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type ecsTable struct {
	sync.RWMutex
	whichType   int
	path        string
	mtime       time.Time
	size        int64
	url         string
	contentHash uint64
	dict  map[string]net.IP
}


func newEcsTable(wtype int, path, url string) *ecsTable {
	return &ecsTable{
		RWMutex: sync.RWMutex{},
		whichType:   wtype,
		path:        path,
		mtime:       time.Time{},
		size:        0,
		url:         url,
		contentHash: 0,
		dict:        make(map[string]net.IP),
	}
}


func (eb *ecsTable) loadFromFile() {
	if eb.whichType != ItemTypePath || len(eb.path) == 0 {
		return
	}
	file, err := os.Open(eb.path)
	if err != nil {
		log.Errorf("ecstable file read error %s", eb.path)
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
	dict, totalLines := eb.parse(file)
	t2 := time.Since(t1)
	log.Debugf("Parsed %v  time spent: %v name added: %v / %v", file.Name(), t2, len(dict), totalLines)

	eb.Lock()
	eb.dict = dict
	eb.mtime = stat.ModTime()
	eb.size = stat.Size()
	eb.Unlock()
}

func (eb *ecsTable) parse(r io.Reader) (map[string]net.IP, uint64) {
	dict := make(map[string]net.IP)
	var totalLines uint64
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		totalLines++

		line := scanner.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}

		attrs := strings.Split(line, ":")
		if len(attrs) != 2 {
			continue
		}
		if !IsIP(attrs[0]) {
			continue
		}
		ip := net.ParseIP(attrs[1])
		if ip != nil {
			dict[attrs[0]] = ip
		}
	}

	return dict, totalLines
}

func (eb *ecsTable) loadFromUrl() {
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

	dict := make(map[string]net.IP)
	var totalLines uint64
	t3 := time.Now()
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		totalLines++
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		attrs := strings.Split(line, ":")
		if len(attrs) != 2 {
			continue
		}
		if !IsIP(attrs[0]) {
			continue
		}
		ip := net.ParseIP(attrs[1])
		if ip != nil {
			dict[attrs[0]] = ip
		}
	}
	t4 := time.Since(t3)
	log.Debugf("Fetched %v, time spent: %v %v, added: %v / %v, hash: %#x",
		eb.url, t2, t4, len(dict), totalLines, contentHash1)

	eb.Lock()
	eb.dict = dict
	eb.contentHash = contentHash1
	eb.Unlock()
}

func (eb *ecsTable) String() string {
	sb := strings.Builder{}
	sb.WriteString("ecsTable:{")
	c := 0
	for k, v := range eb.dict {
		if c >= 5 {
			sb.WriteString("......")
			break
		}
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(v.String())
		sb.WriteString(",")
		c += 1
	}
	sb.WriteString("}")
	return sb.String()
}
