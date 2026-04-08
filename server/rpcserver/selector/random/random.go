package random

import (
	"context"
	"math/rand"

	selector2 "github.com/atliliw/microg/server/rpcserver/selector"
	"github.com/atliliw/microg/server/rpcserver/selector/node/direct"
)

const (
	// Name 是随机负载均衡器名称
	Name = "random"
)

var _ selector2.Balancer = &Balancer{} // Name 是负载均衡器名称

// Balancer 是随机负载均衡器。
type Balancer struct{}

// New 创建一个随机选择器。
func New() selector2.Selector {
	return NewBuilder().Build()
}

// Pick 选择一个加权节点。
func (p *Balancer) Pick(_ context.Context, nodes []selector2.WeightedNode) (selector2.WeightedNode, selector2.DoneFunc, error) {
	if len(nodes) == 0 {
		return nil, nil, selector2.ErrNoAvailable
	}
	cur := rand.Intn(len(nodes))
	selected := nodes[cur]
	d := selected.Pick()
	return selected, d, nil
}

// NewBuilder 返回带有随机负载均衡器的选择器构建器
func NewBuilder() selector2.Builder {
	return &selector2.DefaultBuilder{
		Balancer: &Builder{},
		Node:     &direct.Builder{},
	}
}

// Builder 是随机构建器
type Builder struct{}

// Build 创建负载均衡器
func (b *Builder) Build() selector2.Balancer {
	return &Balancer{}
}
