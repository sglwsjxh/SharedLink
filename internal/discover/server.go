package discover

import (
	"fmt"
	"os"

	"github.com/hashicorp/mdns"
)

type ServiceMeta struct {
	Hostname string
	FileName string
	FileSize int64
}

func Register(port int, meta ServiceMeta) (*mdns.Server, error) {
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	txtRecords := []string{
		"hostname=" + host,
		"filename=" + meta.FileName,
		fmt.Sprintf("filesize=%d", meta.FileSize),
	}

	serviceName := meta.Hostname
	if serviceName == "" {
		serviceName = host
	}

	service, err := mdns.NewMDNSService(
		serviceName,
		"_sharedlink._tcp",
		"",
		"",
		port,
		nil,
		txtRecords,
	)
	if err != nil {
		return nil, fmt.Errorf("create mdns service: %w", err)
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return nil, fmt.Errorf("start mdns server: %w", err)
	}

	return server, nil
}
