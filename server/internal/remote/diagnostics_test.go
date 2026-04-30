package remote

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestDiagnoseFailureClasses(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		checker Checker
		want    string
	}{
		{
			name:    "not configured",
			request: Request{},
			want:    ClassNotConfigured,
		},
		{
			name:    "private route",
			request: Request{PublicURL: "http://192.168.1.20:8097"},
			want:    ClassPrivateRoute,
		},
		{
			name:    "dns failure",
			request: Request{PublicURL: "https://media.example.com"},
			checker: Checker{Resolver: func(context.Context, string) ([]net.IP, error) { return nil, errors.New("no dns") }},
			want:    ClassDNS,
		},
		{
			name:    "nat firewall failure",
			request: Request{PublicURL: "https://media.example.com"},
			checker: Checker{
				Resolver: func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("203.0.113.10")}, nil },
				Dialer:   func(context.Context, string, string) error { return errors.New("timeout") },
			},
			want: ClassNATFirewall,
		},
		{
			name:    "certificate failure",
			request: Request{PublicURL: "https://media.example.com"},
			checker: Checker{
				Resolver: func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("203.0.113.10")}, nil },
				Dialer:   func(context.Context, string, string) error { return nil },
				TLSCheck: func(context.Context, string, string) error { return errors.New("cert") },
			},
			want: ClassCertificate,
		},
		{
			name:    "throughput failure",
			request: Request{PublicURL: "https://media.example.com", MeasuredMbps: 8, RequiredMbps: 25},
			checker: Checker{
				Resolver: func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("203.0.113.10")}, nil },
				Dialer:   func(context.Context, string, string) error { return nil },
				TLSCheck: func(context.Context, string, string) error { return nil },
			},
			want: ClassThroughput,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := tt.checker
			if checker.DialTimeout == 0 {
				checker.DialTimeout = NewChecker().DialTimeout
			}
			got := checker.Diagnose(context.Background(), tt.request, nil)
			if got.FailureClass != tt.want {
				t.Fatalf("expected %q, got %#v", tt.want, got)
			}
			if len(got.NextActions) == 0 {
				t.Fatalf("expected next actions for %#v", got)
			}
		})
	}
}

func TestParseTargetRejectsSensitiveURLParts(t *testing.T) {
	for _, raw := range []string{
		"https://user:pass@example.com",
		"https://example.com/server?token=secret",
		"https://example.com/#token",
	} {
		t.Run(raw, func(t *testing.T) {
			if _, err := ParseTarget(Request{PublicURL: raw}); err == nil {
				t.Fatalf("expected sensitive url to be rejected")
			}
		})
	}
}

func TestReadyDiagnosticSanitizesTarget(t *testing.T) {
	checker := Checker{
		Resolver: func(context.Context, string) ([]net.IP, error) { return []net.IP{net.ParseIP("203.0.113.10")}, nil },
		Dialer:   func(context.Context, string, string) error { return nil },
		TLSCheck: func(context.Context, string, string) error { return nil },
	}
	result := checker.Diagnose(context.Background(), Request{PublicURL: "https://Media.Example.com:9443", RequiredMbps: 10, MeasuredMbps: 40}, nil)
	if result.FailureClass != ClassReady {
		t.Fatalf("expected ready result, got %#v", result)
	}
	if result.Target.Host != "media.example.com" || result.Target.Port != 9443 || result.Target.Scheme != "https" {
		t.Fatalf("unexpected sanitized target: %#v", result.Target)
	}
}
