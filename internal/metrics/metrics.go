package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ActiveWorkers — количество активных воркеров
	// Это кастомная метрика приложения
	ActiveWorkers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_worker_pool_active_workers",
		Help: "Number of currently active workers in the pool",
	})

	// TasksProcessed — количество обработанных задач
	TasksProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_tasks_processed_total",
		Help: "Total number of tasks processed",
	})
)
