package controller

import (
	"context"

	pb "github.com/atliliw/microg/examples/user/api/v1"
	"github.com/atliliw/microg/examples/user/srv/data"
	"github.com/atliliw/microg/examples/user/srv/service"
	"github.com/atliliw/microg/pkg/log"

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
	log.InfoC(ctx, "gRPC请求: GetUserList", log.Uint32("pn", req.GetPn()), log.Uint32("pSize", req.GetPSize()))
	list, err := s.svc.List(ctx, int32(req.GetPn()), int32(req.GetPSize()))
	if err != nil {
		log.ErrorC(ctx, "GetUserList失败", log.Err(err))
		return nil, err
	}

	resp := &pb.UserListResponse{
		Total: int32(list.Total),
		Data:  make([]*pb.UserInfoResponse, 0, len(list.Items)),
	}
	for _, u := range list.Items {
		resp.Data = append(resp.Data, toProto(u))
	}
	log.DebugC(ctx, "GetUserList响应", log.Int32("total", resp.Total), log.Int("count", len(resp.Data)))
	return resp, nil
}

// GetUserByMobile 通过手机号获取用户
func (s *UserServer) GetUserByMobile(ctx context.Context, req *pb.MobileRequest) (*pb.UserInfoResponse, error) {
	log.InfoC(ctx, "gRPC请求: GetUserByMobile", log.String("mobile", req.GetMobile()))
	user, err := s.svc.GetByMobile(ctx, req.GetMobile())
	if err != nil {
		log.ErrorC(ctx, "GetUserByMobile失败", log.Err(err), log.String("mobile", req.GetMobile()))
		return nil, err
	}
	log.DebugC(ctx, "GetUserByMobile响应", log.Int32("id", user.ID))
	return toProto(user), nil
}

// GetUserById 通过ID获取用户
func (s *UserServer) GetUserById(ctx context.Context, req *pb.IdRequest) (*pb.UserInfoResponse, error) {
	log.InfoC(ctx, "gRPC请求: GetUserById", log.Int32("id", req.GetId()))
	user, err := s.svc.GetByID(ctx, req.GetId())
	if err != nil {
		log.ErrorC(ctx, "GetUserById失败", log.Err(err), log.Int32("id", req.GetId()))
		return nil, err
	}
	log.DebugC(ctx, "GetUserById响应", log.Int32("id", user.ID), log.String("nickname", user.NickName))
	return toProto(user), nil
}

// CreateUser 创建用户
func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserInfo) (*pb.UserInfoResponse, error) {
	log.InfoC(ctx, "gRPC请求: CreateUser", log.String("nickname", req.GetNickName()), log.String("mobile", req.GetMobile()))
	user, err := s.svc.Create(ctx, req.GetNickName(), req.GetPassWord(), req.GetMobile())
	if err != nil {
		log.ErrorC(ctx, "CreateUser失败", log.Err(err))
		return nil, err
	}
	log.InfoC(ctx, "CreateUser成功", log.Int32("id", user.ID))
	return toProto(user), nil
}

// UpdateUser 更新用户
func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserInfo) (*emptypb.Empty, error) {
	log.InfoC(ctx, "gRPC请求: UpdateUser", log.Int32("id", req.GetId()), log.String("nickname", req.GetNickName()))
	err := s.svc.Update(ctx, req.GetId(), req.GetNickName(), req.GetGender(), int64(req.GetBirthDay()))
	if err != nil {
		log.ErrorC(ctx, "UpdateUser失败", log.Err(err), log.Int32("id", req.GetId()))
		return nil, err
	}
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