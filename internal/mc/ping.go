package mc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	mcbot "github.com/Tnze/go-mc/bot"
	mcnet "github.com/Tnze/go-mc/net"
)

type PingResult struct {
	Online    bool
	Players   int
	MCTime    time.Duration
	TotalTime time.Duration
}

func Ping(ctx context.Context, addr string, timeout time.Duration) (PingResult, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	totalStart := time.Now() // 总时长计时器

	// resolve address
	resolved, err := ResolveAddr(ctx, addr)
	if err != nil {
		return PingResult{}, fmt.Errorf("resolve: %w", err)
	}
	// establish a TCP connection
	var d net.Dialer
	rawConn, err := d.DialContext(ctx, "tcp", resolved.DialAddr)
	if err != nil {
		return PingResult{}, fmt.Errorf("dial %s: %w", resolved.DialAddr, err)
	}
	// close Nagle
	if tcpConn, ok := rawConn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
	}
	// pack as MC connection
	conn := mcnet.WrapConn(rawConn)
	defer conn.Close()
	// set socket deadline(resolve timeout)
	if deadline, ok := ctx.Deadline(); ok {
		conn.Socket.SetDeadline(deadline)
	}
	// lisen ctx cancel signal
	stop := context.AfterFunc(ctx, func() {
		conn.Close()
	})
	defer stop()
	// call protocol func
	raw, mcDuration, err := ServerListPing(conn, resolved.Host, resolved.Port)
	if err != nil {
		return PingResult{}, fmt.Errorf("server list ping: %w", err)
	}
	// decode json
	var status struct {
		Players struct {
			Online int `json:"online"`
		} `json:"players"`
	}
	if err := json.Unmarshal(raw, &status); err != nil {
		return PingResult{}, fmt.Errorf("json decode: %w", err)
	}

	return PingResult{
		Online:    true,
		Players:   status.Players.Online,
		MCTime:    mcDuration,
		TotalTime: time.Since(totalStart),
	}, nil
}

// legacy go-mc bot pack ping,is not used in this project
//
// Deprecated: use Ping instead.
func BotPing(ctx context.Context, addr string, timeout time.Duration) (bool, int, time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	resp, delay, err := mcbot.PingAndListContext(ctx, addr)
	if err != nil {
		return false, 0, 0, fmt.Errorf("ping %s: %w", addr, err)
	}
	var status struct {
		Players struct {
			Online int `json:"online"`
		} `json:"players"`
	}
	if err := json.Unmarshal(resp, &status); err != nil {
		return false, 0, 0, fmt.Errorf("decode json error: %w", err)
	}
	return true, status.Players.Online, delay, nil
}
