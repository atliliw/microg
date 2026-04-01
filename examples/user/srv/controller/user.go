package controller

import (
	"context"

	pb "microg/examples/user/api/v1"
	"microg/examples/user/srv/data"
	"microg/examples/user/srv/service"

	"google.golang.org/protobuf/types/known/emptypb"
)

// UserServer gRPC 控制器
// 实现 pb.UserServer 接口，处理 gRPC 请求
type UserServer struct {
	pb.UnimplementedUserServer
	svc service.UserService
}

// NewUserServer 创建 gRPC 控制器实例
func NewUserServer(svc service.UserService) pb.UserServer {
	return &UserServer{svc: svc}
}

// GetUserList 获取用户列表
func (s *UserServer) GetUserList(ctx context.Context, req *pb.PageInfo) (*pb.UserListResponse, error) {
	list, err := s.svc.List(ctx, int32(req.GetPn()), int32(req.GetPSize()))
	if err != nil {
		return nil, err
	}

	resp := &pb.UserListResponse{
		Total: int32(list.Total),
		Data:  make([]*pb.UserInfoResponse, 0, len(list.Items)),
	}
	for _, u := range list.Items {
		resp.Data = append(resp.Data, toProto(u))
	}
	return resp, nil
}

// GetUserByMobile 通过手机号获取用户
func (s *UserServer) GetUserByMobile(ctx context.Context, req *pb.MobileRequest) (*pb.UserInfoResponse, error) {
	user, err := s.svc.GetByMobile(ctx, req.GetMobile())
	if err != nil {
		return nil, err
	}
	return toProto(user), nil
}

// GetUserById 通过ID获取用户
func (s *UserServer) GetUserById(ctx context.Context, req *pb.IdRequest) (*pb.UserInfoResponse, error) {
	user, err := s.svc.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return toProto(user), nil
}

// CreateUser 创建用户
func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserInfo) (*pb.UserInfoResponse, error) {
	user, err := s.svc.Create(ctx, req.GetNickName(), req.GetPassWord(), req.GetMobile())
	if err != nil {
		return nil, err
	}
	return toProto(user), nil
}

// UpdateUser 更新用户
func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserInfo) (*emptypb.Empty, error) {
	err := s.svc.Update(ctx, req.GetId(), req.GetNickName(), req.GetGender(), int64(req.GetBirthDay()))
	return &emptypb.Empty{}, err
}

// CheckPassWord 校验密码
func (s *UserServer) CheckPassWord(ctx context.Context, req *pb.PasswordCheckInfo) (*pb.CheckResponse, error) {
	success := s.svc.CheckPassword(req.GetPassword(), req.GetEncryptedPassword())
	return &pb.CheckResponse{Success: success}, nil
}

// toProto 将数据模型转换为 protobuf 响应
func toProto(u *data.User) *pb.UserInfoResponse {
	if u == nil {
		return nil
	}
	return &pb.UserInfoResponse{
		Id:       u.ID,
		PassWord: u.Password,
		Mobile:   u.Mobile,
		NickName: u.NickName,
		Gender:   u.Gender,
		Role:     u.Role,
	}
}

var _ pb.UserServer = (*UserServer)(nil)