package ewma

import (
	"container/list"
	"context"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"

	selector2 "github.com/atliliw/microg/server/rpcserver/selector"
)

const (
	// `cost` 的平均生命周期，经过 Tau*ln(2) 后达到半衰期。
	tau = int64(time.Millisecond * 600)
	// 如果未收集统计数据，我们给端点添加一个大的延迟惩罚
	penalty = uint64(time.Second * 10)
)

var (
	_ selector2.WeightedNode        = &Node{}
	_ selector2.WeightedNodeBuilder = &Builder{}
)

// Node 是端点实例
type Node struct {
	selector2.Node

	// 客户端统计数据
	lag       int64
	success   uint64
	inflight  int64
	inflights *list.List
	// last collected timestamp 上次收集时间戳
	stamp     int64
	predictTs int64
	predict   int64
	// request number in a period time 一段时间内的请求数量
	reqs int64
	// last lastPick timestamp 上次选择时间戳
	lastPick int64

	errHandler func(err error) (isErr bool)
	lk         sync.RWMutex
}

// Builder 是 ewma 节点构建器。
type Builder struct {
	ErrHandler func(err error) (isErr bool)
}

// Build 创建加权节点。
func (b *Builder) Build(n selector2.Node) selector2.WeightedNode {
	s := &Node{
		Node:       n,
		lag:        0,
		success:    1000,
		inflight:   1,
		inflights:  list.New(),
		errHandler: b.ErrHandler,
	}
	return s
}

func (n *Node) health() uint64 {
	return atomic.LoadUint64(&n.success)
}

func (n *Node) load() (load uint64) {
	now := time.Now().UnixNano()
	avgLag := atomic.LoadInt64(&n.lag)
	lastPredictTs := atomic.LoadInt64(&n.predictTs)
	predictInterval := avgLag / 5
	if predictInterval < int64(time.Millisecond*5) {
		predictInterval = int64(time.Millisecond * 5)
	} else if predictInterval > int64(time.Millisecond*200) {
		predictInterval = int64(time.Millisecond * 200)
	}
	if now-lastPredictTs > predictInterval {
		if atomic.CompareAndSwapInt64(&n.predictTs, lastPredictTs, now) {
			var (
				total   int64
				count   int
				predict int64
			)
			n.lk.RLock()
			first := n.inflights.Front()
			for first != nil {
				lag := now - first.Value.(int64)
				if lag > avgLag {
					count++
					total += lag
				}
				first = first.Next()
			}
			if count > (n.inflights.Len()/2 + 1) {
				predict = total / int64(count)
			}
			n.lk.RUnlock()
			atomic.StoreInt64(&n.predict, predict)
		}
	}

	if avgLag == 0 {
		// penalty 是节点刚启动时没有数据时的惩罚值。
		// 默认值是 1e9 * 10
		load = penalty * uint64(atomic.LoadInt64(&n.inflight))
	} else {
		predict := atomic.LoadInt64(&n.predict)
		if predict > avgLag {
			avgLag = predict
		}
		load = uint64(avgLag) * uint64(atomic.LoadInt64(&n.inflight))
	}
	return
}

// Pick 选择节点。
func (n *Node) Pick() selector2.DoneFunc {
	now := time.Now().UnixNano()
	atomic.StoreInt64(&n.lastPick, now)
	atomic.AddInt64(&n.inflight, 1)
	atomic.AddInt64(&n.reqs, 1)
	n.lk.Lock()
	e := n.inflights.PushBack(now)
	n.lk.Unlock()
	return func(ctx context.Context, di selector2.DoneInfo) {
		n.lk.Lock()
		n.inflights.Remove(e)
		n.lk.Unlock()
		atomic.AddInt64(&n.inflight, -1)

		now := time.Now().UnixNano()
		// 获取移动平均比率 w
		stamp := atomic.SwapInt64(&n.stamp, now)
		td := now - stamp
		if td < 0 {
			td = 0
		}
		w := math.Exp(float64(-td) / float64(tau))

		start := e.Value.(int64)
		lag := now - start
		if lag < 0 {
			lag = 0
		}
		oldLag := atomic.LoadInt64(&n.lag)
		if oldLag == 0 {
			w = 0.0
		}
		lag = int64(float64(oldLag)*w + float64(lag)*(1.0-w))
		atomic.StoreInt64(&n.lag, lag)

		success := uint64(1000) // 成功时设为1000，失败时设为0
		if di.Err != nil {
			if n.errHandler != nil {
				if n.errHandler(di.Err) {
					success = 0
				}
			} else {
				// 判断是否为关键错误（超时、取消、服务不可用）
				if errors.Is(di.Err, context.DeadlineExceeded) ||
					errors.Is(di.Err, context.Canceled) {
					success = 0
				}
			}
		}
		oldSuc := atomic.LoadUint64(&n.success)
		success = uint64(float64(oldSuc)*w + float64(success)*(1.0-w))
		atomic.StoreUint64(&n.success, success)
	}
}

// Weight 返回节点有效权重。
func (n *Node) Weight() (weight float64) {
	weight = float64(n.health()*uint64(time.Second)) / float64(n.load())
	return
}

func (n *Node) PickElapsed() time.Duration {
	return time.Duration(time.Now().UnixNano() - atomic.LoadInt64(&n.lastPick))
}

func (n *Node) Raw() selector2.Node {
	return n.Node
}
