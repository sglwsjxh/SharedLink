package discover

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/mdns"
)

type Entry struct {
	Hostname string
	IP       string
	Port     int
	FileName string
	FileSize int64
	Addr     string
}

func Scan(ctx context.Context, timeout time.Duration) ([]Entry, error) {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	entriesCh := make(chan *mdns.ServiceEntry, 50)
	var results []Entry

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	go func() {
		params := mdns.DefaultParams("_sharedlink._tcp")
		params.Entries = entriesCh
		params.Timeout = timeout
		_ = mdns.Query(params)
		close(entriesCh)
	}()

	for entry := range entriesCh {
		if entry == nil {
			continue
		}
		e := parseEntry(entry)
		if e != nil {
			results = append(results, *e)
		}
	}

	return results, nil
}

func parseEntry(entry *mdns.ServiceEntry) *Entry {
	if entry == nil {
		return nil
	}

	ip := ""
	if entry.AddrV4 != nil {
		ip = entry.AddrV4.String()
	} else if entry.AddrV6IPAddr != nil {
		ip = entry.AddrV6IPAddr.String()
	}

	var fileName string
	var fileSize int64
	hostname := entry.Host

	for _, txt := range entry.InfoFields {
		parts := strings.SplitN(txt, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "hostname":
			hostname = val
		case "filename":
			fileName = val
		case "filesize":
			if n, err := strconv.ParseInt(val, 10, 64); err == nil {
				fileSize = n
			}
		}
	}

	return &Entry{
		Hostname: hostname,
		IP:       ip,
		Port:     entry.Port,
		FileName: fileName,
		FileSize: fileSize,
		Addr:     fmt.Sprintf("%s:%d", ip, entry.Port),
	}
}
