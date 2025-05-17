package aggregate

import (
	"github.com/li127103/frp/pkg/metrics/mem"
	"github.com/li127103/frp/pkg/metrics/prometheus"
	"github.com/li127103/frp/server/metrics"
)

// EnableMem start to mark metrics to memory monitor system.
func EnableMem() {
	sm.Add(mem.ServerMetrics)
}

// EnablePrometheus start to mark metrics to prometheus.
func EnablePrometheus() {
	sm.Add(prometheus.ServerMetrics)
}

var sm = &serverMetrics{}

func init() {
	metrics.Register(sm)
}

type serverMetrics struct {
	ms []metrics.ServerMetrics
}

func (m *serverMetrics) Add(sm metrics.ServerMetrics) {
	m.ms = append(m.ms, sm)
}

func (m *serverMetrics) NewClient() {
	for _, v := range m.ms {
		v.NewClient()
	}
}

func (m *serverMetrics) CloseClient() {
	for _, v := range m.ms {
		v.CloseClient()
	}
}

func (m *serverMetrics) NewProxy(name string, proxyType string) {
	for _, v := range m.ms {
		v.NewProxy(name, proxyType)
	}
}

func (m *serverMetrics) CloseProxy(name string, proxyType string) {
	for _, v := range m.ms {
		v.CloseProxy(name, proxyType)
	}
}

func (m *serverMetrics) OpenConnection(name string, proxyType string) {
	for _, v := range m.ms {
		v.OpenConnection(name, proxyType)
	}
}

func (m *serverMetrics) CloseConnection(name string, proxyType string) {
	for _, v := range m.ms {
		v.CloseConnection(name, proxyType)
	}
}

func (m *serverMetrics) AddTrafficIn(name string, proxyType string, trafficBytes int64) {
	for _, v := range m.ms {
		v.AddTrafficIn(name, proxyType, trafficBytes)
	}
}

func (m *serverMetrics) AddTrafficOut(name string, proxyType string, trafficBytes int64) {
	for _, v := range m.ms {
		v.AddTrafficOut(name, proxyType, trafficBytes)
	}
}
