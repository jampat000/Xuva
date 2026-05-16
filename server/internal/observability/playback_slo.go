package observability

import (
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type PlaybackSLOMetrics struct {
	Window                         string  `json:"window"`
	SessionsStarted                int64   `json:"sessionsStarted"`
	SessionsWithFirstHeartbeat     int64   `json:"sessionsWithFirstHeartbeat"`
	StartupP95Ms                   float64 `json:"startupP95Ms"`
	StartupMaxMs                   float64 `json:"startupMaxMs"`
	StartupUnder5sRate             float64 `json:"startupUnder5sRate"`
	RouteChanges                   int64   `json:"routeChanges"`
	AdaptiveStalls                 int64   `json:"adaptiveStalls"`
	AdaptiveRecoveries             int64   `json:"adaptiveRecoveries"`
	AdaptiveRecoveredWithin10sRate float64 `json:"adaptiveRecoveredWithin10sRate"`
}

type playbackSLOState struct {
	startedSessions     map[string]time.Time
	stallStarted        map[string]time.Time
	sessionsStarted     int64
	firstHeartbeats     int64
	routeChanges        int64
	adaptiveStalls      int64
	adaptiveRecoveries  int64
	startupLatenciesMS  []float64
	recoveryLatenciesMS []float64
}

func newPlaybackSLOState() playbackSLOState {
	return playbackSLOState{
		startedSessions:     map[string]time.Time{},
		stallStarted:        map[string]time.Time{},
		startupLatenciesMS:  make([]float64, 0, 512),
		recoveryLatenciesMS: make([]float64, 0, 512),
	}
}

func (s *Service) observePlaybackEventLocked(eventType string, data any, createdAt time.Time) {
	switch eventType {
	case "session.started":
		sessionID := asStringField(data, "id", "sessionId")
		if sessionID != "" {
			s.playback.startedSessions[sessionID] = createdAt
		}
		s.playback.sessionsStarted++
	case "session.updated":
		sessionID := asStringField(data, "id", "sessionId")
		if sessionID == "" {
			return
		}
		startedAt, ok := s.playback.startedSessions[sessionID]
		if !ok {
			return
		}
		latency := createdAt.Sub(startedAt).Seconds() * 1000
		if latency < 0 {
			latency = 0
		}
		s.playback.firstHeartbeats++
		s.playback.startupLatenciesMS = appendCapped(s.playback.startupLatenciesMS, latency, 512)
		delete(s.playback.startedSessions, sessionID)
	case "session.stopped", "session.stale":
		sessionID := asStringField(data, "id", "sessionId")
		if sessionID != "" {
			delete(s.playback.startedSessions, sessionID)
		}
	case "session.route.changed":
		s.playback.routeChanges++
	case "adaptive.stall":
		sessionID := asStringField(data, "sessionId", "sessionID", "id")
		s.playback.adaptiveStalls++
		if sessionID != "" {
			s.playback.stallStarted[sessionID] = createdAt
		}
	case "adaptive.recovered":
		sessionID := asStringField(data, "sessionId", "sessionID", "id")
		s.playback.adaptiveRecoveries++
		if sessionID != "" {
			if startedAt, ok := s.playback.stallStarted[sessionID]; ok {
				recovery := createdAt.Sub(startedAt).Seconds() * 1000
				if recovery < 0 {
					recovery = 0
				}
				s.playback.recoveryLatenciesMS = appendCapped(s.playback.recoveryLatenciesMS, recovery, 512)
				delete(s.playback.stallStarted, sessionID)
			}
		}
	}
}

func (p playbackSLOState) snapshot() PlaybackSLOMetrics {
	startupP95 := percentile(p.startupLatenciesMS, 95)
	startupMax := percentile(p.startupLatenciesMS, 100)
	startupUnder5sRate := ratioUnder(p.startupLatenciesMS, 5000)
	recoveredUnder10sRate := ratioUnder(p.recoveryLatenciesMS, 10000)
	return PlaybackSLOMetrics{
		Window:                         "process-lifetime",
		SessionsStarted:                p.sessionsStarted,
		SessionsWithFirstHeartbeat:     p.firstHeartbeats,
		StartupP95Ms:                   startupP95,
		StartupMaxMs:                   startupMax,
		StartupUnder5sRate:             startupUnder5sRate,
		RouteChanges:                   p.routeChanges,
		AdaptiveStalls:                 p.adaptiveStalls,
		AdaptiveRecoveries:             p.adaptiveRecoveries,
		AdaptiveRecoveredWithin10sRate: recoveredUnder10sRate,
	}
}

func EvaluatePlaybackSLOAlerts(metrics PlaybackSLOMetrics) []Alert {
	alerts := []Alert{}
	if metrics.SessionsWithFirstHeartbeat >= 5 && metrics.StartupUnder5sRate < 0.95 {
		alerts = append(alerts, Alert{
			Severity: "warning",
			Code:     "playback_startup_slo",
			Message:  "Playback startup SLO degraded: fewer than 95% of sessions reached first heartbeat within 5s.",
			Action:   "Reduce background load, inspect transcode queue saturation, and validate network path/client profile routing.",
		})
	}
	if metrics.SessionsWithFirstHeartbeat >= 5 && metrics.StartupP95Ms > 8000 {
		alerts = append(alerts, Alert{
			Severity: "warning",
			Code:     "playback_startup_p95",
			Message:  "Playback startup p95 is above 8s.",
			Action:   "Inspect /api/metrics queue pressure and recent playback route changes, then lower concurrent heavy work.",
		})
	}
	if metrics.AdaptiveRecoveries >= 5 && metrics.AdaptiveRecoveredWithin10sRate < 0.9 {
		alerts = append(alerts, Alert{
			Severity: "warning",
			Code:     "adaptive_recovery_slo",
			Message:  "Adaptive recovery SLO degraded: fewer than 90% of stalls recovered within 10s.",
			Action:   "Lower remote quality limits and verify throughput guidance for impacted remote clients.",
		})
	}
	return alerts
}

func appendCapped(input []float64, value float64, max int) []float64 {
	input = append(input, value)
	if len(input) > max {
		input = input[len(input)-max:]
	}
	return input
}

func percentile(input []float64, p float64) float64 {
	if len(input) == 0 {
		return 0
	}
	if p <= 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	data := append([]float64(nil), input...)
	sort.Float64s(data)
	index := int((p / 100) * float64(len(data)-1))
	return data[index]
}

func ratioUnder(input []float64, threshold float64) float64 {
	if len(input) == 0 {
		return 1
	}
	var pass int
	for _, value := range input {
		if value <= threshold {
			pass++
		}
	}
	return float64(pass) / float64(len(input))
}

func asStringField(data any, keys ...string) string {
	switch typed := data.(type) {
	case map[string]any:
		for _, key := range keys {
			if value := strings.TrimSpace(toString(typed[key])); value != "" {
				return value
			}
		}
	default:
		if data == nil {
			return ""
		}
		raw, err := json.Marshal(data)
		if err != nil {
			return ""
		}
		var parsed map[string]any
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return ""
		}
		for _, key := range keys {
			if value := strings.TrimSpace(toString(parsed[key])); value != "" {
				return value
			}
		}
	}
	return ""
}

func toString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}
