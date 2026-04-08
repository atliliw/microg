package selector

import (
	"strconv"

	"github.com/atliliw/microg/registry"
)

// DefaultNode 是选择器节点
type DefaultNode struct {
	scheme   string
	addr     string
	weight   *int64
	version  string
	name     string
	metadata map[string]string
}

// Scheme 返回节点协议
func (n *DefaultNode) Scheme() string {
	return n.scheme
}

// Address 返回节点地址
func (n *DefaultNode) Address() string {
	return n.addr
}

// ServiceName 返回节点服务名称
func (n *DefaultNode) ServiceName() string {
	return n.name
}

// InitialWeight 返回节点初始权重
func (n *DefaultNode) InitialWeight() *int64 {
	return n.weight
}

// Version 返回节点版本
func (n *DefaultNode) Version() string {
	return n.version
}

// Metadata 返回节点元数据
func (n *DefaultNode) Metadata() map[string]string {
	return n.metadata
}

// NewNode 创建新节点
func NewNode(scheme, addr string, ins *registry.ServiceInstance) Node {
	n := &DefaultNode{
		scheme: scheme,
		addr:   addr,
	}
	if ins != nil {
		n.name = ins.Name
		n.version = ins.Version
		n.metadata = ins.Metadata
		if str, ok := ins.Metadata["weight"]; ok {
			if weight, err := strconv.ParseInt(str, 10, 64); err == nil {
				n.weight = &weight
			}
		}
	}
	return n
}
