package main

import (
	"github.com/DataDog/datadog-go/statsd"
	"github.com/buildkite/buildkite-metrics/collector"
)

// StatsD sends metrics to StatsD (Datadog spec)
type StatsD struct {
	client *statsd.Client
}

func NewStatsDClient(host string) (*StatsD, error) {
	c, err := statsd.NewBuffered(host, 100)
	if err != nil {
		return nil, err
	}
	// prefix every metric with the app name
	c.Namespace = "buildkite."
	return &StatsD{
		client: c,
	}, nil
}

func (cb *StatsD) Collect(r *collector.Result) error {
	for name, value := range r.Totals {
		if err := cb.client.Gauge(name, float64(value), []string{}, 1.0); err != nil {
			return err
		}
	}

	for queue, counts := range r.Queues {
		for name, value := range counts {
			if err := cb.client.Gauge("queues."+name, float64(value), []string{"queue:" + queue}, 1.0); err != nil {
				return err
			}
		}
	}

	for pipeline, counts := range r.Pipelines {
		for name, value := range counts {
			if err := cb.client.Gauge("pipelines."+name, float64(value), []string{"pipeline:" + pipeline}, 1.0); err != nil {
				return err
			}
		}
	}

	return nil
}
