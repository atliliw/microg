package selector

import (
	"context"
	"sync/atomic"
)

// Default 是组合选择器。
type Default struct {
	NodeBuilder WeightedNodeBuilder
	Balancer    Balancer

	nodes atomic.Value
}

// Select 选择一个节点。
func (d *Default) Select(ctx context.Context) (selected Node, done DoneFunc, err error) {
	var (
		candidates []WeightedNode
	)
	nodes, ok := d.nodes.Load().([]WeightedNode)
	if !ok {
		return nil, nil, ErrNoAvailable
	}
	candidates = nodes

	if len(candidates) == 0 {
		return nil, nil, ErrNoAvailable
	}
	wn, done, err := d.Balancer.Pick(ctx, candidates)
	if err != nil {
		return nil, nil, err
	}
	p, ok := FromPeerContext(ctx)
	if ok {
		p.Node = wn.Raw()
	}
	return wn.Raw(), done, nil
}

// Apply 更新节点信息。
func (d *Default) Apply(nodes []Node) {
	weightedNodes := make([]WeightedNode, 0, len(nodes))
	for _, n := range nodes {
		weightedNodes = append(weightedNodes, d.NodeBuilder.Build(n))
	}
	// TODO: 不要删除未变更的节点
	d.nodes.Store(weightedNodes)
}

// DefaultBuilder 是默认选择器构建器。
type DefaultBuilder struct {
	Node     WeightedNodeBuilder
	Balancer BalancerBuilder
}

// Build 创建选择器
func (db *DefaultBuilder) Build() Selector {
	return &Default{
		NodeBuilder: db.Node,
		Balancer:    db.Balancer.Build(),
	}
}
