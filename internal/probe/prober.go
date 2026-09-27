package probe

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"

	"mc-monitor/internal/mc"
)

type Target struct {
	Name string
	Addr string
}
type Prober struct {
	targets  []Target
	cache    *Cache
	interval time.Duration
	timeout  time.Duration
	limit    int
	logger   *slog.Logger
}

func NewProber(targets []Target, cache *Cache, interval time.Duration, timeout time.Duration, limit int, logger *slog.Logger) *Prober {
	return &Prober{
		targets:  targets,
		cache:    cache,
		interval: interval,
		timeout:  timeout,
		limit:    limit,
		logger:   logger,
	}
}

func (p *Prober) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	p.probeAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.probeAll(ctx)
		}
	}
}

func (p *Prober) probeAll(ctx context.Context) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(p.limit)

	for _, t := range p.targets {
		g.Go(func() error {
			online, players, rtt, err := mc.Ping(ctx, t.Addr, p.timeout)
			p.cache.Set(t.Name, ProbeResult{
				Online:        online,
				Players:       players,
				Timestamp:     time.Now(),
				Err:           err,
				ProbeDuration: rtt,
			})
			return nil
		})
	}
	g.Wait()
}
