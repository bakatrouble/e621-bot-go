package utils

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	reg *prometheus.Registry

	postsChecked prometheus.Counter
	postsMatched prometheus.Counter
	postsSent    prometheus.Counter
}

func (m *Metrics) IncPostsChecked() {
	m.postsChecked.Inc()
}

func (m *Metrics) IncPostsMatched() {
	m.postsMatched.Inc()
}

func (m *Metrics) IncPostsSent() {
	m.postsSent.Inc()
}

func (m *Metrics) PromHandler() http.Handler {
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{Registry: m.reg})
}

func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		reg: reg,
		postsChecked: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "posts_checked",
			Help: "Number of posts checked",
		}),
		postsMatched: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "posts_matched",
			Help: "Number of posts matched",
		}),
		postsSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "posts_sent",
			Help: "Number of posts sent",
		}),
	}
	reg.MustRegister(m.postsChecked)
	reg.MustRegister(m.postsMatched)
	reg.MustRegister(m.postsSent)
	return m
}
