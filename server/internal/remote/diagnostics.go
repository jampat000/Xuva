package remote

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	ClassReady         = "ready"
	ClassNotConfigured = "not_configured"
	ClassPrivateRoute  = "private_route"
	ClassDNS           = "dns"
	ClassNATFirewall   = "nat_firewall"
	ClassCertificate   = "certificate"
	ClassThroughput    = "throughput"
	ClassInvalidInput  = "invalid_input"
)

type Request struct {
	PublicURL     string  `json:"publicUrl"`
	ExpectedPort  int     `json:"expectedPort"`
	RequiredMbps  float64 `json:"requiredMbps"`
	MeasuredMbps  float64 `json:"measuredMbps"`
	SkipNetworkIO bool    `json:"skipNetworkIo"`
	AssumeHTTPS   bool    `json:"assumeHttps"`
}

type Result struct {
	Status       string   `json:"status"`
	FailureClass string   `json:"failureClass"`
	Route        string   `json:"route"`
	Summary      string   `json:"summary"`
	Target       Target   `json:"target"`
	Checks       []Check  `json:"checks"`
	NextActions  []string `json:"nextActions"`
	Privacy      []string `json:"privacy"`
}

type Target struct {
	Scheme string `json:"scheme,omitempty"`
	Host   string `json:"host,omitempty"`
	Port   int    `json:"port,omitempty"`
}

type Check struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Checker struct {
	DialTimeout time.Duration
	Resolver    func(context.Context, string) ([]net.IP, error)
	Dialer      func(context.Context, string, string) error
	TLSCheck    func(context.Context, string, string) error
}

func NewChecker() Checker {
	return Checker{
		DialTimeout: 4 * time.Second,
		Resolver: func(ctx context.Context, host string) ([]net.IP, error) {
			return net.DefaultResolver.LookupIP(ctx, "ip", host)
		},
		Dialer: func(ctx context.Context, network string, address string) error {
			dialer := net.Dialer{}
			conn, err := dialer.DialContext(ctx, network, address)
			if err != nil {
				return err
			}
			return conn.Close()
		},
		TLSCheck: func(ctx context.Context, host string, address string) error {
			dialer := tls.Dialer{Config: &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}}
			conn, err := dialer.DialContext(ctx, "tcp", address)
			if err != nil {
				return err
			}
			return conn.Close()
		},
	}
}

func (c Checker) Diagnose(ctx context.Context, request Request, lanAddresses []string) Result {
	result := Result{
		Status:       "needs_action",
		FailureClass: ClassNotConfigured,
		Route:        "not_configured",
		Summary:      "Remote access has not been configured yet.",
		NextActions:  Guidance(ClassNotConfigured),
		Privacy: []string{
			"Diagnostics only return scheme, host, and port.",
			"URL paths, query strings, usernames, passwords, and tokens are not stored or returned.",
			"Vyrden does not use a vendor relay for this check.",
		},
	}
	if strings.TrimSpace(request.PublicURL) == "" {
		result.Checks = append(result.Checks, Check{Name: "Public address", Status: "warn", Code: ClassNotConfigured, Message: "Add the public URL, VPN name, or reverse proxy address you expect remote players to use."})
		return result
	}

	target, err := ParseTarget(request)
	if err != nil {
		result.FailureClass = ClassInvalidInput
		result.Route = "invalid"
		result.Summary = "The remote address could not be understood."
		result.NextActions = Guidance(ClassInvalidInput)
		result.Checks = append(result.Checks, Check{Name: "Public address", Status: "fail", Code: ClassInvalidInput, Message: "Enter a URL such as https://media.example.com or https://example.com:8097."})
		return result
	}
	result.Target = target
	result.Route = classifyRoute(target.Host, lanAddresses)
	if result.Route == "private" {
		result.FailureClass = ClassPrivateRoute
		result.Summary = "This address looks private. It should work only on your LAN or inside your VPN/mesh network."
		result.NextActions = Guidance(ClassPrivateRoute)
		result.Checks = append(result.Checks, Check{Name: "Route type", Status: "warn", Code: ClassPrivateRoute, Message: "Use this route for VPN/mesh access, or enter a public reverse proxy/domain for internet access."})
		return withThroughput(result, request)
	}

	result.Checks = append(result.Checks, Check{Name: "Route type", Status: "pass", Code: "public_route", Message: "The address is not a private LAN address."})
	if request.SkipNetworkIO {
		result.Status = "ready"
		result.FailureClass = ClassReady
		result.Summary = "Remote route format is valid. Live DNS/TCP/TLS checks were skipped."
		result.NextActions = Guidance(ClassReady)
		return withThroughput(result, request)
	}

	timeout := c.DialTimeout
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ips, err := c.resolve(checkCtx, target.Host)
	if err != nil || len(ips) == 0 {
		result.FailureClass = ClassDNS
		result.Summary = "The remote name does not resolve to an address."
		result.NextActions = Guidance(ClassDNS)
		result.Checks = append(result.Checks, Check{Name: "DNS", Status: "fail", Code: ClassDNS, Message: "Create or fix the DNS record for this hostname."})
		return result
	}
	result.Checks = append(result.Checks, Check{Name: "DNS", Status: "pass", Code: "dns_ok", Message: "The hostname resolves."})

	address := net.JoinHostPort(target.Host, strconv.Itoa(target.Port))
	if err := c.dial(checkCtx, "tcp", address); err != nil {
		result.FailureClass = ClassNATFirewall
		result.Summary = "Vyrden could not open a TCP connection to the remote address."
		result.NextActions = Guidance(ClassNATFirewall)
		result.Checks = append(result.Checks, Check{Name: "Port reachability", Status: "fail", Code: ClassNATFirewall, Message: "Check port forwarding, router firewall, ISP CGNAT, and host firewall rules."})
		return result
	}
	result.Checks = append(result.Checks, Check{Name: "Port reachability", Status: "pass", Code: "tcp_ok", Message: "The port accepted a TCP connection."})

	if target.Scheme == "https" || request.AssumeHTTPS {
		if err := c.tlsCheck(checkCtx, target.Host, address); err != nil {
			result.FailureClass = ClassCertificate
			result.Summary = "The HTTPS route answered, but the certificate check failed."
			result.NextActions = Guidance(ClassCertificate)
			result.Checks = append(result.Checks, Check{Name: "TLS certificate", Status: "fail", Code: ClassCertificate, Message: "Use a valid certificate for the hostname remote players will use."})
			return result
		}
		result.Checks = append(result.Checks, Check{Name: "TLS certificate", Status: "pass", Code: "tls_ok", Message: "The certificate is valid for this hostname."})
	}

	result.Status = "ready"
	result.FailureClass = ClassReady
	result.Summary = "Remote route checks passed."
	result.NextActions = Guidance(ClassReady)
	return withThroughput(result, request)
}

func ParseTarget(request Request) (Target, error) {
	raw := strings.TrimSpace(request.PublicURL)
	if raw == "" {
		return Target{}, errors.New("empty url")
	}
	if !strings.Contains(raw, "://") {
		if request.AssumeHTTPS {
			raw = "https://" + raw
		} else {
			raw = "http://" + raw
		}
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return Target{}, errors.New("invalid url")
	}
	host := parsed.Hostname()
	if host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return Target{}, errors.New("unsafe url")
	}
	port := request.ExpectedPort
	if parsed.Port() != "" {
		parsedPort, err := strconv.Atoi(parsed.Port())
		if err != nil {
			return Target{}, errors.New("invalid port")
		}
		port = parsedPort
	}
	if port == 0 {
		if parsed.Scheme == "https" {
			port = 443
		} else {
			port = 80
		}
	}
	if port < 1 || port > 65535 {
		return Target{}, errors.New("invalid port")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return Target{}, errors.New("unsupported scheme")
	}
	return Target{Scheme: scheme, Host: strings.ToLower(host), Port: port}, nil
}

func Guidance(class string) []string {
	switch class {
	case ClassReady:
		return []string{"Use this URL when pairing a remote player.", "Keep HTTPS enabled for public internet access.", "If playback buffers, lower the remote quality limit or use optimized versions."}
	case ClassPrivateRoute:
		return []string{"Use this address only on LAN, VPN, or a mesh network.", "For public internet access, configure a reverse proxy or port forward with HTTPS.", "Do not expose Vyrden directly without authentication and TLS."}
	case ClassDNS:
		return []string{"Create an A/AAAA or CNAME record for the hostname.", "If your public IP changes, configure dynamic DNS.", "Wait for DNS propagation, then run diagnostics again."}
	case ClassNATFirewall:
		return []string{"Forward the chosen external port to the Vyrden server or reverse proxy.", "Allow the port through the Windows/Linux firewall.", "If your ISP uses CGNAT, use a VPN/mesh network or reverse proxy tunnel you control."}
	case ClassCertificate:
		return []string{"Install a valid certificate for the hostname.", "Check reverse proxy TLS settings and certificate renewal.", "Avoid self-signed certificates for TV/mobile clients unless the device explicitly trusts them."}
	case ClassThroughput:
		return []string{"Lower the remote quality limit.", "Use adaptive streaming when available.", "Create an optimized version for unreliable links."}
	case ClassInvalidInput:
		return []string{"Enter only the base remote address, for example https://media.example.com.", "Remove usernames, passwords, tokens, paths, and query strings.", "Use ports between 1 and 65535."}
	default:
		return []string{"Review the failed check and run diagnostics again after changing your network setup."}
	}
}

func withThroughput(result Result, request Request) Result {
	if request.RequiredMbps > 0 && request.MeasuredMbps > 0 && request.MeasuredMbps < request.RequiredMbps {
		result.Status = "needs_action"
		result.FailureClass = ClassThroughput
		result.Summary = "The route is reachable, but measured throughput is below the selected playback target."
		result.NextActions = Guidance(ClassThroughput)
		result.Checks = append(result.Checks, Check{Name: "Throughput", Status: "warn", Code: ClassThroughput, Message: "Use a lower remote quality limit or an optimized version for this connection."})
		return result
	}
	if request.RequiredMbps > 0 && request.MeasuredMbps > 0 {
		result.Checks = append(result.Checks, Check{Name: "Throughput", Status: "pass", Code: "throughput_ok", Message: "Measured throughput meets the selected target."})
	}
	return result
}

func (c Checker) resolve(ctx context.Context, host string) ([]net.IP, error) {
	if c.Resolver == nil {
		c = NewChecker()
	}
	return c.Resolver(ctx, host)
}

func (c Checker) dial(ctx context.Context, network string, address string) error {
	if c.Dialer == nil {
		c = NewChecker()
	}
	return c.Dialer(ctx, network, address)
}

func (c Checker) tlsCheck(ctx context.Context, host string, address string) error {
	if c.TLSCheck == nil {
		c = NewChecker()
	}
	return c.TLSCheck(ctx, host, address)
}

func classifyRoute(host string, lanAddresses []string) string {
	ip := net.ParseIP(host)
	if ip != nil && isPrivateIP(ip) {
		return "private"
	}
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "localhost" || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".lan") {
		return "private"
	}
	for _, address := range lanAddresses {
		parsed, err := url.Parse(address)
		if err != nil {
			continue
		}
		if strings.EqualFold(parsed.Hostname(), host) {
			return "private"
		}
	}
	return "public"
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
		return true
	}
	return false
}
