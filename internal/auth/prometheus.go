package auth

import "github.com/prometheus/client_golang/prometheus"

// Define your metrics
var (
	loginAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_login_attempts_total",
			Help: "Total number of login attempts.",
		},
		[]string{"status"},
	)
	registerAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_register_attempts_total",
			Help: "Total number of registration attempts.",
		},
		[]string{"status"},
	)
	refreshAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_refresh_attempts_total",
			Help: "Total number of token refresh attempts.",
		},
		[]string{"status"},
	)
)
