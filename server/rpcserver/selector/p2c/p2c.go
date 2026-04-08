package p2c

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	selector2 "github.com/atliliw/microg/server/rpcserver/selector"
	"github.com/atliliw/microg/server/rpcserver/selector/node/ewma"
)

const (
	forcePick = time.Second * 3
	// Name 是负载均衡器名称
	Name = "p2c"
)

var _ selector2.Balancer = &Balancer{}

// New 创建一个 p2c 选择器。
func New() selector2.Selector {
	return NewBuilder().Build()
}

// Balancer 是 p2c 选择器。
type Balancer struct {
	mu     sync.Mutex
	r      *rand.Rand
	picked int64
}

// 选择两个不同的节点。
func (s *Balancer) prePick(nodes []selector2.WeightedNode) (nodeA selector2.WeightedNode, nodeB selector2.WeightedNode) {
	s.mu.Lock()
	a := s.r.Intn(len(nodes))
	b := s.r.Intn(len(nodes) - 1)
	s.mu.Unlock()
	if b >= a {
		b = b + 1
	}
	nodeA, nodeB = nodes[a], nodes[b]
	return
}

// Pick 选择一个节点。
func (s *Balancer) Pick(ctx context.Context, nodes []selector2.WeightedNode) (selector2.WeightedNode, selector2.DoneFunc, error) {
	if len(nodes) == 0 {
		return nil, nil, selector2.ErrNoAvailable
	}
	if len(nodes) == 1 {
		done := nodes[0].Pick()
		return nodes[0], done, nil
	}

	var pc, upc selector2.WeightedNode
	nodeA, nodeB := s.prePick(nodes)
	// meta.Weight 是服务发布者在服务发现中设置的权重
	if nodeB.Weight() > nodeA.Weight() {
		pc, upc = nodeB, nodeA
	} else {
		pc, upc = nodeA, nodeB
	}

	// 如果失败节点在 forceGap 时间内从未被选择过，则强制选择一次
	// 利用强制机会触发成功率和延迟的更新
	if upc.PickElapsed() > forcePick && atomic.CompareAndSwapInt64(&s.picked, 0, 1) {
		pc = upc
		atomic.StoreInt64(&s.picked, 0)
	}
	done := pc.Pick()
	return pc, done, nil
}

// NewBuilder 返回带有 p2c 负载均衡器的选择器构建器
func NewBuilder() selector2.Builder {
	return &selector2.DefaultBuilder{
		Balancer: &Builder{},
		Node:     &ewma.Builder{},
	}
}

// Builder 是 p2c 构建器
type Builder struct{}

// Build 创建负载均衡器
func (b *Builder) Build() selector2.Balancer {
	return &Balancer{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}
