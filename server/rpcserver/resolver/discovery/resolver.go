package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"time"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"

	"github.com/atliliw/microg/pkg/log"
	"github.com/atliliw/microg/registry"
)

// discoveryResolver 实现 resolver.Resolver 接口
// 负责监听服务变化并更新 gRPC 连接地址
type discoveryResolver struct {
	w  registry.Watcher    // 服务监听器
	cc resolver.ClientConn // gRPC 客户端连接

	ctx    context.Context    // 上下文
	cancel context.CancelFunc // 取消函数

	insecure bool // 是否不安全连接
}

// watch 监听服务变化
// 这是一个阻塞方法，在后台 goroutine 中运行
func (r *discoveryResolver) watch() {
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}

		// 阻塞等待服务变化
		ins, err := r.w.Next()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Errorf("[resolver] Failed to watch discovery endpoint: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// 更新连接地址
		r.update(ins)
	}
}

// update 更新 gRPC 客户端连接地址
// 将服务实例列表转换为 resolver.Address 并通知 gRPC
func (r *discoveryResolver) update(ins []*registry.ServiceInstance) {
	addrs := make([]resolver.Address, 0)
	endpoints := make(map[string]struct{})

	for _, in := range ins {
		// 解析端点地址，提取 grpc 协议的地址
		endpoint, err := ParseEndpoint(in.Endpoints, "grpc", !r.insecure)
		if err != nil {
			log.Errorf("[resolver] Failed to parse discovery endpoint: %v", err)
			continue
		}
		if endpoint == "" {
			continue
		}

		// 过滤重复端点
		if _, ok := endpoints[endpoint]; ok {
			continue
		}
		endpoints[endpoint] = struct{}{}

		// 构建 resolver.Address
		addr := resolver.Address{
			ServerName: in.Name,
			Attributes: parseAttributes(in.Metadata),
			Addr:       endpoint,
		}
		addr.Attributes = addr.Attributes.WithValue("rawServiceInstance", in)
		addrs = append(addrs, addr)
	}

	if len(addrs) == 0 {
		log.Warnf("[resolver] Zero endpoint found,refused to write, instances: %v", ins)
		return
	}

	// 更新 gRPC 客户端状态，触发重新连接
	err := r.cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
		log.Errorf("[resolver] failed to update state: %s", err)
	}

	b, _ := json.Marshal(ins)
	log.Infof("[resolver] update instances: %s", b)
}

// Close 关闭解析器
// 停止监听并释放资源
func (r *discoveryResolver) Close() {
	r.cancel()
	err := r.w.Stop()
	if err != nil {
		log.Errorf("[resolver] failed to watch top: %s", err)
	}
}

// ResolveNow 立即解析
// discovery 解析器通过 watch 实时更新，此方法为空实现
func (r *discoveryResolver) ResolveNow(options resolver.ResolveNowOptions) {}

// parseAttributes 将元数据转换为 gRPC attributes
func parseAttributes(md map[string]string) *attributes.Attributes {
	var a *attributes.Attributes
	for k, v := range md {
		if a == nil {
			a = attributes.New(k, v)
		} else {
			a = a.WithValue(k, v)
		}
	}
	return a
}

// NewEndpoint 创建端点 URL
// 参数:
//   - scheme: 协议 (grpc, http)
//   - host: 主机地址 (ip:port)
//   - isSecure: 是否安全连接
func NewEndpoint(scheme, host string, isSecure bool) *url.URL {
	var query string
	if isSecure {
		query = "isSecure=true"
	}
	return &url.URL{Scheme: scheme, Host: host, RawQuery: query}
}

// ParseEndpoint 从端点列表中解析指定协议的地址
// 参数:
//   - endpoints: 端点列表，如 ["grpc://192.168.1.10:9001", "http://192.168.1.10:8080"]
//   - scheme: 目标协议，如 "grpc"
//   - isSecure: 是否安全连接
//
// 返回:
//   - string: 解析出的地址，如 "192.168.1.10:9001"
//   - error: 错误信息
func ParseEndpoint(endpoints []string, scheme string, isSecure bool) (string, error) {
	for _, e := range endpoints {
		u, err := url.Parse(e)
		if err != nil {
			return "", err
		}
		if u.Scheme == scheme {
			if IsSecure(u) == isSecure {
				return u.Host, nil
			}
		}
	}
	return "", nil
}

// IsSecure 从 URL 中解析 isSecure 参数
func IsSecure(u *url.URL) bool {
	ok, err := strconv.ParseBool(u.Query().Get("isSecure"))
	if err != nil {
		return false
	}
	return ok
}
