package metrics

import (
	"log/slog"
	"time"
)

type Cleaner struct {
	m                         *Metrics
	removeWhenInactiveMinutes int
}

func NewCleaner(m *Metrics, removeWhenInactiveMinutes int) *Cleaner {
	return &Cleaner{m, removeWhenInactiveMinutes}
}

func (c *Cleaner) Start() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.m.pm.lock.Lock()
				for source, sourceMetrics := range c.m.pm.metrics {
					if time.Since(sourceMetrics.lastAccessed) > time.Duration(c.removeWhenInactiveMinutes)*time.Minute {
						delete(c.m.pm.metrics, source)
						slog.Info("removed inactive source", "source", source)
					}
				}
				c.m.pm.lock.Unlock()
			}
		}
	}()
	slog.Debug("scheduled metrics cleanup")
}
