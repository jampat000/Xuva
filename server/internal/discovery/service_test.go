package discovery

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/jampat000/Xuva/server/internal/config"
)

type testAnnouncer struct {
	shutdowns int
}

func (a *testAnnouncer) Shutdown() error {
	a.shutdowns++
	return nil
}

func TestServiceDisabledByConfig(t *testing.T) {
	service := NewServiceForTest(config.Config{
		ServerName:       "Family Room",
		HTTPAddr:         "0.0.0.0:8097",
		DiscoveryEnabled: false,
	}, func(AdvertiseConfig) (announcer, error) {
		t.Fatalf("factory should not be called when discovery is disabled")
		return nil, nil
	})

	service.Start(context.Background())
	status := service.Status()

	if status.Enabled {
		t.Fatalf("expected disabled status, got %#v", status)
	}
	if status.Running {
		t.Fatalf("expected discovery not running, got %#v", status)
	}
	if status.Note == "" {
		t.Fatalf("expected plain disabled note, got %#v", status)
	}
}

func TestServiceCapturesExpectedAdvertiseConfig(t *testing.T) {
	var captured AdvertiseConfig
	service := NewServiceForTest(config.Config{
		ServerName:           "Family Room",
		HTTPAddr:             "192.168.1.20:8097",
		DiscoveryEnabled:     true,
		DiscoveryServiceType: DefaultServiceType,
	}, func(cfg AdvertiseConfig) (announcer, error) {
		captured = cfg
		return &testAnnouncer{}, nil
	})

	service.Start(context.Background())
	status := service.Status()

	if !status.Running {
		t.Fatalf("expected running discovery status, got %#v", status)
	}
	if captured.ServiceName != "Family Room" {
		t.Fatalf("expected service name from server name, got %#v", captured)
	}
	if captured.ServiceType != DefaultServiceType {
		t.Fatalf("expected service type %q, got %#v", DefaultServiceType, captured)
	}
	if captured.Port != 8097 {
		t.Fatalf("expected advertised port 8097, got %#v", captured)
	}
	if len(captured.IPs) != 1 || !captured.IPs[0].Equal(net.ParseIP("192.168.1.20")) {
		t.Fatalf("expected advertised IP 192.168.1.20, got %#v", captured.IPs)
	}
	if len(captured.TXTRecords) == 0 {
		t.Fatalf("expected safe txt records, got %#v", captured)
	}
	if !hasTXTRecord(captured.TXTRecords, "serverName=Family Room") {
		t.Fatalf("expected Xuva display name in TXT records, got %#v", captured.TXTRecords)
	}
	if !hasTXTPrefix(captured.TXTRecords, "hostName=") {
		t.Fatalf("expected network host name in TXT records, got %#v", captured.TXTRecords)
	}
	if !hasTXTPrefix(captured.TXTRecords, "web=http://") {
		t.Fatalf("expected derived web origin in TXT records, got %#v", captured.TXTRecords)
	}
	if status.HostName == "" || status.WebURL == "" {
		t.Fatalf("expected network fields in status, got %#v", status)
	}
}

func TestServiceUsesConfiguredCanonicalWebOrigin(t *testing.T) {
	var captured AdvertiseConfig
	service := NewServiceForTest(config.Config{
		ServerName:           "Family Room",
		HTTPAddr:             "192.168.1.20:8097",
		CanonicalWebOrigin:   "http://media.example.test:8097",
		DiscoveryEnabled:     true,
		DiscoveryServiceType: DefaultServiceType,
	}, func(cfg AdvertiseConfig) (announcer, error) {
		captured = cfg
		return &testAnnouncer{}, nil
	})

	service.Start(context.Background())
	status := service.Status()

	if status.WebURL != "http://media.example.test:8097" {
		t.Fatalf("expected configured canonical web origin in status, got %#v", status)
	}
	if !hasTXTRecord(captured.TXTRecords, "web=http://media.example.test:8097") {
		t.Fatalf("expected configured canonical web origin in TXT records, got %#v", captured.TXTRecords)
	}
}

func TestServiceFallsBackToXuvaName(t *testing.T) {
	var captured AdvertiseConfig
	service := NewServiceForTest(config.Config{
		ServerName:           "   ",
		HTTPAddr:             "192.168.1.20:8097",
		DiscoveryEnabled:     true,
		DiscoveryServiceType: "",
	}, func(cfg AdvertiseConfig) (announcer, error) {
		captured = cfg
		return &testAnnouncer{}, nil
	})

	service.Start(context.Background())

	if captured.ServiceName != "Xuva" {
		t.Fatalf("expected fallback service name Xuva, got %#v", captured)
	}
	if captured.ServiceType != DefaultServiceType {
		t.Fatalf("expected default service type, got %#v", captured)
	}
}

func TestServiceFailureDoesNotCrashAndExposesPlainStatus(t *testing.T) {
	service := NewServiceForTest(config.Config{
		ServerName:           "Family Room",
		HTTPAddr:             "192.168.1.20:8097",
		DiscoveryEnabled:     true,
		DiscoveryServiceType: DefaultServiceType,
	}, func(AdvertiseConfig) (announcer, error) {
		return nil, errors.New("multicast socket unavailable")
	})

	service.Start(context.Background())
	status := service.Status()

	if status.Running {
		t.Fatalf("expected failed discovery to remain stopped, got %#v", status)
	}
	if status.LastError != "multicast socket unavailable" {
		t.Fatalf("expected plain last error, got %#v", status)
	}
	if status.Note == "" {
		t.Fatalf("expected user-facing note after failure, got %#v", status)
	}
}

func TestServiceDoesNotAdvertiseLoopbackOnlyBind(t *testing.T) {
	service := NewServiceForTest(config.Config{
		ServerName:           "Desk Xuva",
		HTTPAddr:             "127.0.0.1:8097",
		DiscoveryEnabled:     true,
		DiscoveryServiceType: DefaultServiceType,
	}, func(AdvertiseConfig) (announcer, error) {
		t.Fatalf("factory should not be called for loopback-only bind")
		return nil, nil
	})

	service.Start(context.Background())
	status := service.Status()

	if status.Running {
		t.Fatalf("expected loopback-only bind to stay stopped, got %#v", status)
	}
	if status.LastError != "" {
		t.Fatalf("expected no hard failure for loopback-only bind, got %#v", status)
	}
	if status.Note == "" {
		t.Fatalf("expected explanatory note for loopback-only bind, got %#v", status)
	}
}

func hasTXTRecord(records []string, expected string) bool {
	for _, record := range records {
		if record == expected {
			return true
		}
	}
	return false
}

func hasTXTPrefix(records []string, prefix string) bool {
	for _, record := range records {
		if strings.HasPrefix(record, prefix) {
			return true
		}
	}
	return false
}
