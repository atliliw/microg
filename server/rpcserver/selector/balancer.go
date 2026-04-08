package selector

import (
	"context"
	"time"
)

// Balancer 是负载均衡器接口
type Balancer interface {
	Pick(ctx context.Context, nodes []WeightedNode) (selected WeightedNode, done DoneFunc, err error)
}

// BalancerBuilder 构建负载均衡器
type BalancerBuilder interface {
	Build() Balancer
}

// WeightedNode 实时计算调度权重
type WeightedNode interface {
	Node

	// Raw 返回原始节点
	Raw() Node

	// Weight 返回运行时计算的权重
	Weight() float64

	// Pick 选择节点
	Pick() DoneFunc

	// PickElapsed 返回自上次选择以来经过的时间
	PickElapsed() time.Duration
}

// WeightedNodeBuilder 是加权节点构建器
type WeightedNodeBuilder interface {
	Build(Node) WeightedNode
}
