package pictures

import "github.com/prometheus/client_golang/prometheus"

// Define your metrics
var (
	getPicturesRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pictures_get_requests_total",
			Help: "Total number of get pictures requests.",
		},
		[]string{"status"},
	)
	getPictureRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pictures_single_get_requests_total",
			Help: "Total number of get single picture requests.",
		},
		[]string{"status"},
	)
	getTotalPicturesRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pictures_total_get_requests_total",
			Help: "Total number of get total pictures requests.",
		},
		[]string{"status"},
	)

	uploadPictureRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pictures_upload_requests_total",
			Help: "Total number of upload picture requests.",
		},
		[]string{"status"},
	)
)
