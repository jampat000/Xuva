package discovery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/mdns"

	"github.com/jampat000/Xuva/server/internal/config"
)

const (
	DefaultServiceType = "_xuva._tcp"
	serviceDomain      = "local."
	bootstrapAPIPath   = "/api/client/bootstrap"
)

type Status struct {
	Enabled     bool     `json:"enabled"`
	Running     bool     `json:"running"`
	ServiceName string   `json:"serviceName,omitempty"`
	ServiceType string   `json:"serviceType,omitempty"`
	Port        int      `json:"port,omitempty"`
	TXTRecords  []string `json:"txtRecords,omitempty"`
	LastError   string   `json:"lastError,omitempty"`
	Note        string   `json:"note,omitempty"`
}

type AdvertiseConfig struct {
	ServiceName string
	ServiceType string
	Port        int
	IPs         []net.IP
	TXTRecords  []string
}

type announcer interface {
	Shutdown() error
}

type announcerFactory func(AdvertiseConfig) (announcer, error)

type Service struct {
	cfg     config.Config
	factory announcerFactory

	mu        sync.RWMutex
	announcer announcer
	status    Status
	stopOnce  sync.Once
}

func NewService(cfg config.Config) *Service {
	return &Service{
		cfg:     cfg,
		factory: newMDNSAnnouncer,
		status: Status{
			Enabled:     cfg.DiscoveryEnabled,
			ServiceName: displayServiceName(cfg.ServerName),
			ServiceType: fullServiceType(cfg.DiscoveryServiceType),
		},
	}
}

func NewServiceForTest(cfg config.Config, factory func(AdvertiseConfig) (announcer, error)) *Service {
	service := NewService(cfg)
	service.factory = factory
	return service
}

func (s *Service) Start(ctx context.Context) {
	status := Status{
		Enabled:     s.cfg.DiscoveryEnabled,
		ServiceName: displayServiceName(s.cfg.ServerName),
		ServiceType: fullServiceType(s.cfg.DiscoveryServiceType),
	}
	if !s.cfg.DiscoveryEnabled {
		status.Note = "Local discovery is turned off."
		s.setStatus(status)
		slog.Info("Local discovery disabled")
		return
	}

	port, err := httpPort(s.cfg.HTTPAddr)
	if err != nil {
		status.LastError = "http port is unavailable"
		status.Note = "Local discovery is not running."
		s.setStatus(status)
		slog.Warn("Local discovery could not start: http port is unavailable")
		return
	}
	status.Port = port

	if config.HTTPAddrLoopbackOnly(s.cfg.HTTPAddr) {
		status.Note = "This server is listening only on this device right now."
		s.setStatus(status)
		slog.Info("Local discovery disabled", "reason", "server is listening only on loopback")
		return
	}

	ips, err := advertisedIPs(s.cfg.HTTPAddr)
	if err != nil {
		status.LastError = err.Error()
		status.Note = "Local discovery is not running."
		s.setStatus(status)
		slog.Warn("Local discovery could not start: " + err.Error())
		return
	}

	txtRecords := []string{
		"app=xuva",
		"api=" + bootstrapAPIPath,
		"serverName=" + status.ServiceName,
	}
	sort.Strings(txtRecords)
	status.TXTRecords = append([]string(nil), txtRecords...)

	instance := AdvertiseConfig{
		ServiceName: status.ServiceName,
		ServiceType: normalizedServiceType(s.cfg.DiscoveryServiceType),
		Port:        port,
		IPs:         ips,
		TXTRecords:  txtRecords,
	}
	announcer, err := s.factory(instance)
	if err != nil {
		status.LastError = err.Error()
		status.Note = "Local discovery is not running."
		s.setStatus(status)
		slog.Warn("Local discovery could not start: " + err.Error())
		return
	}

	status.Running = true
	s.mu.Lock()
	s.announcer = announcer
	s.status = status
	s.mu.Unlock()
	slog.Info("Local discovery started as " + status.ServiceName)

	go func() {
		<-ctx.Done()
		s.Shutdown()
	}()
}

func (s *Service) Shutdown() {
	s.stopOnce.Do(func() {
		s.mu.Lock()
		announcer := s.announcer
		s.announcer = nil
		if s.status.Running {
			s.status.Running = false
			if s.status.Note == "" {
				s.status.Note = "Local discovery is not running."
			}
		}
		s.mu.Unlock()

		if announcer != nil {
			if err := announcer.Shutdown(); err != nil {
				slog.Warn("Local discovery could not stop cleanly: " + err.Error())
			} else {
				slog.Info("Local discovery stopped")
			}
		}
	})
}

func (s *Service) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	status := s.status
	if len(status.TXTRecords) > 0 {
		status.TXTRecords = append([]string(nil), status.TXTRecords...)
	}
	return status
}

func (s *Service) setStatus(status Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

func newMDNSAnnouncer(cfg AdvertiseConfig) (announcer, error) {
	service, err := mdns.NewMDNSService(
		cfg.ServiceName,
		cfg.ServiceType,
		serviceDomain,
		localHostRecord(),
		cfg.Port,
		cfg.IPs,
		cfg.TXTRecords,
	)
	if err != nil {
		return nil, err
	}
	return mdns.NewServer(&mdns.Config{Zone: service})
}

func displayServiceName(value string) string {
	normalized, err := config.NormalizeServerName(value)
	if err != nil {
		return "Xuva"
	}
	return normalized
}

func localHostRecord() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	if strings.HasSuffix(trimmed, ".") {
		return trimmed
	}
	if strings.Contains(trimmed, ".") {
		return trimmed + "."
	}
	return trimmed + ".local."
}

func normalizedServiceType(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return DefaultServiceType
	}
	trimmed = strings.TrimSuffix(trimmed, ".local.")
	trimmed = strings.TrimSuffix(trimmed, ".local")
	return trimmed
}

func fullServiceType(value string) string {
	return normalizedServiceType(value) + ".local."
}

func httpPort(httpAddr string) (int, error) {
	host := strings.TrimSpace(httpAddr)
	if host == "" {
		return 0, errors.New("http port is unavailable")
	}
	_, port, err := net.SplitHostPort(host)
	if err != nil {
		return 0, errors.New("http port is unavailable")
	}
	parsed, err := strconv.Atoi(port)
	if err != nil || parsed < 1 || parsed > 65535 {
		return 0, errors.New("http port is unavailable")
	}
	return parsed, nil
}

func advertisedIPs(httpAddr string) ([]net.IP, error) {
	host, _, err := net.SplitHostPort(strings.TrimSpace(httpAddr))
	if err != nil {
		return nil, errors.New("http address is invalid")
	}
	host = strings.TrimSpace(strings.Trim(host, "[]"))
	switch {
	case host == "", host == "0.0.0.0", host == "::":
		return activeLANIPs()
	case strings.EqualFold(host, "localhost"):
		return nil, errors.New("server is listening only on this device")
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() {
			return nil, errors.New("server is listening only on this device")
		}
		return []net.IP{ip}, nil
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("could not resolve advertised address")
	}
	filtered := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if ip == nil || ip.IsLoopback() {
			continue
		}
		filtered = append(filtered, ip)
	}
	if len(filtered) == 0 {
		return nil, errors.New("no local network address is available")
	}
	return dedupeIPs(filtered), nil
}

func activeLANIPs() ([]net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, errors.New("no local network address is available")
	}
	var output []net.IP
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch value := addr.(type) {
			case *net.IPNet:
				ip = value.IP
			case *net.IPAddr:
				ip = value.IP
			}
			if ip == nil || ip.IsLoopback() || ip.IsUnspecified() {
				continue
			}
			output = append(output, ip)
		}
	}
	output = dedupeIPs(output)
	if len(output) == 0 {
		return nil, errors.New("no local network address is available")
	}
	return output, nil
}

func dedupeIPs(values []net.IP) []net.IP {
	seen := map[string]bool{}
	output := make([]net.IP, 0, len(values))
	for _, ip := range values {
		if ip == nil {
			continue
		}
		normalized := ip.String()
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		output = append(output, ip)
	}
	sort.Slice(output, func(i int, j int) bool {
		return output[i].String() < output[j].String()
	})
	return output
}
