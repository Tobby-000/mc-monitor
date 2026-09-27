package mc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const DefaultPort uint16 = 25565

type ResolvedAddr struct {
	DialAddr string // 实际用于 TCP 拨号的地址，可能来自 SRV 记录
	Host     string // 握手包中填写的 Server Address，即用户最初输入的域名。在某些情况下使用，比如服务器会用这个值作为反代后的不同子后端的依据。
	Port     uint16 // 握手包中填写的 Server Port，实际连接的端口
}

func ResolveAddr(ctx context.Context, addr string) (ResolvedAddr, error) {
	if addr == "" {
		return ResolvedAddr{}, fmt.Errorf("empty address")
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		var addrErr *net.AddrError
		if errors.As(err, &addrErr) && strings.Contains(addrErr.Err, "missing port") {
			// SRV resolve
			_, srvs, srvErr := net.DefaultResolver.LookupSRV(ctx, "minecraft", "tcp", addr)
			if srvErr == nil && len(srvs) > 0 {
				target := strings.TrimSuffix(srvs[0].Target, ".")
				return ResolvedAddr{
					DialAddr: net.JoinHostPort(target, strconv.Itoa(int(srvs[0].Port))),
					Host:     addr,
					Port:     srvs[0].Port,
				}, nil
			}
			return ResolvedAddr{
				DialAddr: net.JoinHostPort(strings.Trim(addr, "[]"), strconv.Itoa(int(DefaultPort))),
				Host:     addr,
				Port:     DefaultPort,
			}, nil
		}
		return ResolvedAddr{}, fmt.Errorf("split addr %s: %w", addr, err)
	}
	if host == "" {
		return ResolvedAddr{}, fmt.Errorf("empty host in %q", addr)
	}
	p, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return ResolvedAddr{}, fmt.Errorf("parse port %s: %w", port, err)
	}
	return ResolvedAddr{
		DialAddr: addr,
		Host:     host,
		Port:     uint16(p),
	}, nil
}
