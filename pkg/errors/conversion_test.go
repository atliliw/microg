package errors

import (
	"testing"

	"google.golang.org/grpc/status"
)

func TestGrpcErrorConversion(t *testing.T) {
	Register(defaultCoder{201004, 400, "Invalid mobile number", ""})

	err := WithCode(201004, "%s", "查询用户总数失败")
	t.Logf("原始错误: %+v, code: %d", err, getErrorCode(err))

	grpcErr := ToGrpcError(err)
	t.Logf("转换为 gRPC 错误: %v", grpcErr)

	st, _ := status.FromError(grpcErr)
	t.Logf("gRPC status: code=%d, message=%s", st.Code(), st.Message())

	backErr := FromGrpcError(grpcErr)
	t.Logf("转换回来的错误: %+v, code: %d", backErr, getErrorCode(backErr))

	coder := ParseCoder(backErr)
	t.Logf("解析后的 Coder: code=%d, msg=%s", coder.Code(), coder.String())

	if coder.Code() != 201004 {
		t.Errorf("期望 code=201004, 实际 code=%d", coder.Code())
	}
}

func getErrorCode(e error) int {
	if v, ok := e.(*withCode); ok {
		return v.code
	}
	return 0
}
