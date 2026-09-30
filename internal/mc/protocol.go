package mc

import (
	"fmt"
	"time"

	"github.com/Tnze/go-mc/data/packetid"
	mcnet "github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
)

const protocolVersion = -1

// ServerListPing 在已建立的连接上执行 Server List Ping 流程，
// 返回 Status Response 里的原始 JSON 字节，以及整个协议交互的耗时。
//
// 返回的 duration 覆盖：写 Handshake、写 Status Request、读 Status Response。
// 它不包含 TCP 握手和 DNS 解析，也不等于纯网络 RTT —— 它包含了服务端
// 处理请求、序列化 JSON（含 favicon）的时间，更接近玩家打开服务器列表的体验。
// 如需纯网络 RTT，请另发 PING/PONG 包单独测量。
func ServerListPing(conn *mcnet.Conn, host string, port uint16) ([]byte, time.Duration, error) {
	// handshack
	if err := conn.WritePacket(pk.Marshal(
		0x00, //handshake pack
		pk.VarInt(protocolVersion),
		pk.String(host),
		pk.UnsignedShort(port),
		pk.Byte(1), //status mode
	)); err != nil {
		return nil, 0, fmt.Errorf("write handshake: %w", err)
	}
	// send status request
	if err := conn.WritePacket(pk.Marshal(
		packetid.ServerboundStatusRequest,
	)); err != nil {
		return nil, 0, fmt.Errorf("write status request: %w", err)
	}
	// status response
	var p pk.Packet
	if err := conn.ReadPacket(&p); err != nil {
		return nil, 0, fmt.Errorf("read status response: %w", err)
	}
	// packid check
	wantID := int32(packetid.ClientboundStatusResponse)
	if p.ID != wantID {
		return nil, 0, fmt.Errorf("unexpected packet id: %d, want %d", p.ID, wantID)
	}
	var raw pk.String
	// scan response packet to get json
	if err := p.Scan(&raw); err != nil {
		return nil, 0, fmt.Errorf("scan status response: %w", err)
	}
	// ping-pong
	pingStart := time.Now()
	send:=pingStart.UnixMilli()
	if err:=conn.WritePacket(pk.Marshal(
		packetid.ServerboundStatusPingRequest,
		pk.Long(send),
	));err!=nil{
		return nil, 0, fmt.Errorf("write ping request: %w", err)
	}
	var pong pk.Packet
	if err:=conn.ReadPacket(&pong);err!=nil{
		return nil, 0, fmt.Errorf("read pong response: %w", err)
	}
	rtt:=time.Since(pingStart)
	wantPongID:=int32(packetid.ClientboundStatusPongResponse)
	if pong.ID!=wantPongID{
		return nil, 0, fmt.Errorf("unexpected packet id: %d, want %d", pong.ID, wantPongID)
	}
	var echoed pk.Long
	if err := pong.Scan(&echoed); err != nil {
		return nil, 0, fmt.Errorf("scan pong response: %w", err)
	}
	_ = echoed
	return []byte(raw),rtt, nil
}
