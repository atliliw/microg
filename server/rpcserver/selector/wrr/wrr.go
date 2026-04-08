package wrr

import (
	"context"
	"sync"

	selector2 "github.com/atliliw/microg/server/rpcserver/selector"
	"github.com/atliliw/microg/server/rpcserver/selector/node/direct"
)

const (
	// Name 是加权轮询负载均衡器名称
	Name = "wrr"
)

var _ selector2.Balancer = &Balancer{} // Name 是负载均衡器名称

// Option 是随机构建器选项。

// Balancer 是加权轮询负载均衡器。
type Balancer struct {
	mu            sync.Mutex
	currentWeight map[string]float64
}

// New 创建一个加权轮询选择器。
func New() selector2.Selector {
	return NewBuilder().Build()
}

// Pick 选择一个加权节点。
func (p *Balancer) Pick(_ context.Context, nodes []selector2.WeightedNode) (selector2.WeightedNode, selector2.DoneFunc, error) {
	if len(nodes) == 0 {
		return nil, nil, selector2.ErrNoAvailable
	}
	var totalWeight float64
	var selected selector2.WeightedNode
	var selectWeight float64

	// nginx 加权轮询负载均衡算法: http://blog.csdn.net/zhangskd/article/details/50194069
	p.mu.Lock()
	for _, node := range nodes {
		totalWeight += node.Weight()
		cwt := p.currentWeight[node.Address()]
		// 当前权重 += 有效权重
		cwt += node.Weight()
		p.currentWeight[node.Address()] = cwt
		if selected == nil || selectWeight < cwt {
			selectWeight = cwt
			selected = node
		}
	}
	p.currentWeight[selected.Address()] = selectWeight - totalWeight
	p.mu.Unlock()

	d := selected.Pick()
	return selected, d, nil
}

// NewBuilder 返回带有加权轮询负载均衡器的选择器构建器
func NewBuilder() selector2.Builder {
	return &selector2.DefaultBuilder{
		Balancer: &Builder{},
		Node:     &direct.Builder{},
	}
}

// Builder 是加权轮询构建器
type Builder struct{}

// Build 创建负载均衡器
func (b *Builder) Build() selector2.Balancer {
	return &Balancer{currentWeight: make(map[string]float64)}
}
