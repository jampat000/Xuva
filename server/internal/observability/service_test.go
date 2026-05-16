package observability

import "testing"

func TestEvaluateAlertsDetectsSaturatedQueue(t *testing.T) {
	alerts := EvaluateAlerts([]map[string]any{
		{"name": "transcode", "workers": 1, "active": int64(1), "queued": 2},
	}, nil)
	if len(alerts) != 1 {
		t.Fatalf("expected one alert, got %#v", alerts)
	}
	if alerts[0].Code != "queue_saturated" || alerts[0].Action == "" {
		t.Fatalf("expected actionable queue alert, got %#v", alerts[0])
	}
}

func TestEvaluateAlertsDetectsAPIErrorRate(t *testing.T) {
	alerts := EvaluateAlerts(nil, []RequestMetric{
		{Method: "GET", Path: "/api/example", Count: 5, Errors: 1, LastCorrelation: "req-test"},
	})
	if len(alerts) != 1 {
		t.Fatalf("expected one alert, got %#v", alerts)
	}
	if alerts[0].Code != "api_error_rate" || alerts[0].Action == "" {
		t.Fatalf("expected actionable api alert, got %#v", alerts[0])
	}
}

func TestEvaluatePlaybackSLOAlertsDetectsStartupRegression(t *testing.T) {
	alerts := EvaluatePlaybackSLOAlerts(PlaybackSLOMetrics{
		SessionsWithFirstHeartbeat: 10,
		StartupUnder5sRate:         0.8,
		StartupP95Ms:               9200,
	})
	if len(alerts) < 2 {
		t.Fatalf("expected startup alerts, got %#v", alerts)
	}
}
