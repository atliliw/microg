package direct

import (
	"context"
	"sync/atomic"
	"time"

	selector2 "github.com/atliliw/microg/server/rpcserver/selector"
)

const (
	defaultWeight = 100
)

var (
	_ selector2.WeightedNode        = &Node{}
	_ selector2.WeightedNodeBuilder = &Builder{}
)

// Node 是端点实例
type Node struct {
	selector2.Node

	// lastPick 上次选择时间戳
	lastPick int64
}

// Builder 是直连节点构建器
type Builder struct{}

// Build 创建节点
func (*Builder) Build(n selector2.Node) selector2.WeightedNode {
	return &Node{Node: n, lastPick: 0}
}

func (n *Node) Pick() selector2.DoneFunc {
	now := time.Now().UnixNano()
	atomic.StoreInt64(&n.lastPick, now)
	return func(ctx context.Context, di selector2.DoneInfo) {}
}

// Weight 返回节点有效权重
func (n *Node) Weight() float64 {
	if n.InitialWeight() != nil {
		return float64(*n.InitialWeight())
	}
	return defaultWeight
}

func (n *Node) PickElapsed() time.Duration {
	return time.Duration(time.Now().UnixNano() - atomic.LoadInt64(&n.lastPick))
}

func (n *Node) Raw() selector2.Node {
	return n.Node
}
