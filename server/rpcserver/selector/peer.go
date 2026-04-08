package selector

import (
	"context"
)

type peerKey struct{}

// Peer 包含 RPC 调用对端的信息，如地址和认证信息。
type Peer struct {
	// Node 是对端节点。
	Node Node
}

// NewPeerContext 创建一个带有对端信息的新上下文。
func NewPeerContext(ctx context.Context, p *Peer) context.Context {
	return context.WithValue(ctx, peerKey{}, p)
}

// FromPeerContext 从上下文中返回对端信息（如果存在）。
func FromPeerContext(ctx context.Context) (p *Peer, ok bool) {
	p, ok = ctx.Value(peerKey{}).(*Peer)
	return
}
