package resources

type WorkloadClass string

const (
	PlaybackCritical WorkloadClass = "playback_critical"
	Interactive      WorkloadClass = "interactive"
	Background       WorkloadClass = "background"
)

type Limits struct {
	ScanWorkers      int `json:"scanWorkers"`
	ProbeWorkers     int `json:"probeWorkers"`
	TranscodeWorkers int `json:"transcodeWorkers"`
	GPUWorkers       int `json:"gpuWorkers"`
}

type Manager struct {
	limits Limits
}

func NewManager(limits Limits) *Manager {
	return &Manager{limits: limits}
}

func (m *Manager) Limits() Limits {
	return m.limits
}

func (m *Manager) Classes() []map[string]any {
	return []map[string]any{
		{"name": PlaybackCritical, "priority": 100, "examples": []string{"streaming", "active_transcode", "subtitle_delivery", "session_heartbeat"}},
		{"name": Interactive, "priority": 50, "examples": []string{"api", "dashboard", "users", "devices"}},
		{"name": Background, "priority": 10, "examples": []string{"scan", "probe", "metadata", "intro_detection", "cleanup"}},
	}
}
