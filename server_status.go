package router

import (
	"sync"
)

type HttpMetrics map[string]*HttpMetric

type ServerStatus struct {
	// TODO: Need to copy/paste all fileds in Metric here
	// workaround for json package doesn't support anonymous fields
	HttpMetric
	sync.Mutex

	Urls           int                    `json:"urls"`
	Droplets       int                    `json:"droplets"`
	RequestsPerSec int                    `json:"requests_per_sec"`
	Tags           map[string]HttpMetrics `json:"tags"`
}

type HttpMetric struct {
	Requests     int           `json:"requests"`
	Latency      *Distribution `json:"latency"`
	Responses2xx int           `json:"responses_2xx"`
	Responses3xx int           `json:"responses_3xx"`
	Responses4xx int           `json:"responses_4xx"`
	Responses5xx int           `json:"responses_5xx"`
	ResponsesXxx int           `json:"responses_xxx"`
}

func NewServerStatus() *ServerStatus {
	s := new(ServerStatus)

	s.Tags = make(map[string]HttpMetrics)

	s.Latency = NewDistribution(s, "overall")
	s.Latency.Reset()

	tags := []string{"component", "framework", "runtime"}
	for _, tag := range tags {
		s.Tags[tag] = make(HttpMetrics)
	}

	return s
}

func NewHttpMetric(name string) *HttpMetric {
	m := new(HttpMetric)

	m.Latency = NewDistribution(m, name)
	m.Latency.Reset()

	return m
}

func (s *ServerStatus) IncRequests() {
	s.Lock()
	defer s.Unlock()

	s.Requests++
}

func (s *ServerStatus) IncDroplets() {
	s.Lock()
	defer s.Unlock()

	s.Droplets++
}

func (s *ServerStatus) RecordResponse(status int, latency int, tags map[string]string) {
	if latency < 0 {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.record(status, latency)

	for key, value := range tags {
		if s.Tags[key] == nil {
			continue
		}

		if s.Tags[key][value] == nil {
			s.Tags[key][value] = NewHttpMetric(key + "." + value)
		}
		s.Tags[key][value].record(status, latency)
	}
}

func (m *HttpMetric) record(status int, latency int) {
	if status >= 200 && status < 300 {
		m.Responses2xx++
	} else if status >= 300 && status < 400 {
		m.Responses3xx++
	} else if status >= 400 && status < 500 {
		m.Responses4xx++
	} else if status >= 500 && status < 600 {
		m.Responses5xx++
	} else {
		m.ResponsesXxx++
	}

	m.Latency.Add(int64(latency))
}