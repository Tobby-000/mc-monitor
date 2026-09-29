package probe

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"mcping-exporter/internal/mc"
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
	// running 防止上一轮探测未完成时，下一轮 Ticker 触发导致探测重叠。
	running  atomic.Bool
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
	// 当上一次探测还在运行时，跳过本轮
	if p.running.Load() {
		p.logger.Warn("previous probe still running, skipping this round")
		return
	}
	p.running.Store(true)
	defer p.running.Store(false)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(p.limit)

	for _, t := range p.targets {
		g.Go(func() error {
			res, err := mc.Ping(ctx, t.Addr, p.timeout)
			p.cache.Set(t.Name, ProbeResult{
				Online:           res.Online && err == nil,
				Players:          res.Players,
				Timestamp:        time.Now(),
				Err:              err,
				ProtocolDuration: res.MCTime,
				TotalDuration:    res.TotalTime,
			})
			return nil
		})
	}
	g.Wait()
}
