package data

import (
	"context"

	pb "github.com/atliliw/microg/examples/user/api/v1"
	"github.com/atliliw/microg/pkg/log"
)

// User 用户数据模型（客户端）
type User struct {
	ID       int32
	Mobile   string
	Password string
	NickName string
	Gender   string
	Role     int32
}

// UserList 用户列表结构
type UserList struct {
	Total int32
	Items []*User
}

// UserRepo 用户数据仓库接口
// 封装对 user-srv gRPC 服务的调用
type UserRepo interface {
	List(ctx context.Context, page, pageSize int32) (*UserList, error)
	GetByID(ctx context.Context, id int32) (*User, error)
	GetByMobile(ctx context.Context, mobile string) (*User, error)
	Create(ctx context.Context, nickName, password, mobile string) (*User, error)
}

type userRepo struct {
	client pb.UserClient
}

// NewUserRepo 创建用户数据仓库实例
func NewUserRepo(client pb.UserClient) UserRepo {
	return &userRepo{client: client}
}

// List 分页查询用户列表
func (r *userRepo) List(ctx context.Context, page, pageSize int32) (*UserList, error) {
	log.DebugC(ctx, "gRPC调用: GetUserList", log.Int32("page", page), log.Int32("pageSize", pageSize))
	resp, err := r.client.GetUserList(ctx, &pb.PageInfo{
		Pn:    uint32(page),
		PSize: uint32(pageSize),
	})
	if err != nil {
		log.ErrorC(ctx, "gRPC调用失败: GetUserList", log.Err(err))
		return nil, err
	}

	list := &UserList{
		Total: resp.Total,
		Items: make([]*User, 0, len(resp.Data)),
	}
	for _, u := range resp.Data {
		list.Items = append(list.Items, &User{
			ID:       u.Id,
			Mobile:   u.Mobile,
			Password: u.PassWord,
			NickName: u.NickName,
			Gender:   u.Gender,
			Role:     u.Role,
		})
	}
	log.DebugC(ctx, "gRPC调用成功: GetUserList", log.Int32("total", list.Total), log.Int("count", len(list.Items)))
	return list, nil
}

// GetByID 通过ID查询用户
func (r *userRepo) GetByID(ctx context.Context, id int32) (*User, error) {
	log.DebugC(ctx, "gRPC调用: GetUserById", log.Int32("id", id))
	resp, err := r.client.GetUserById(ctx, &pb.IdRequest{Id: id})
	if err != nil {
		log.ErrorC(ctx, "gRPC调用失败: GetUserById", log.Err(err), log.Int32("id", id))
		return nil, err
	}
	log.DebugC(ctx, "gRPC调用成功: GetUserById", log.Int32("id", resp.Id), log.String("nickname", resp.NickName))
	return &User{
		ID:       resp.Id,
		Mobile:   resp.Mobile,
		Password: resp.PassWord,
		NickName: resp.NickName,
		Gender:   resp.Gender,
		Role:     resp.Role,
	}, nil
}

// GetByMobile 通过手机号查询用户
func (r *userRepo) GetByMobile(ctx context.Context, mobile string) (*User, error) {
	log.DebugC(ctx, "gRPC调用: GetUserByMobile", log.String("mobile", mobile))
	resp, err := r.client.GetUserByMobile(ctx, &pb.MobileRequest{Mobile: mobile})
	if err != nil {
		log.ErrorC(ctx, "gRPC调用失败: GetUserByMobile", log.Err(err), log.String("mobile", mobile))
		return nil, err
	}
	log.DebugC(ctx, "gRPC调用成功: GetUserByMobile", log.Int32("id", resp.Id))
	return &User{
		ID:       resp.Id,
		Mobile:   resp.Mobile,
		Password: resp.PassWord,
		NickName: resp.NickName,
		Gender:   resp.Gender,
		Role:     resp.Role,
	}, nil
}

// Create 创建用户
func (r *userRepo) Create(ctx context.Context, nickName, password, mobile string) (*User, error) {
	log.InfoC(ctx, "gRPC调用: CreateUser", log.String("nickname", nickName), log.String("mobile", mobile))
	resp, err := r.client.CreateUser(ctx, &pb.CreateUserInfo{
		NickName: nickName,
		PassWord: password,
		Mobile:   mobile,
	})
	if err != nil {
		log.ErrorC(ctx, "gRPC调用失败: CreateUser", log.Err(err))
		return nil, err
	}
	log.InfoC(ctx, "gRPC调用成功: CreateUser", log.Int32("id", resp.Id), log.String("nickname", resp.NickName))
	return &User{
		ID:       resp.Id,
		Mobile:   resp.Mobile,
		Password: resp.PassWord,
		NickName: resp.NickName,
	}, nil
}

var _ UserRepo = (*userRepo)(nil)
