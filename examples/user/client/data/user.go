package data

import (
	"context"

	pb "microg/examples/user/api/v1"
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
	resp, err := r.client.GetUserList(ctx, &pb.PageInfo{
		Pn:    uint32(page),
		PSize: uint32(pageSize),
	})
	if err != nil {
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
	return list, nil
}

// GetByID 通过ID查询用户
func (r *userRepo) GetByID(ctx context.Context, id int32) (*User, error) {
	resp, err := r.client.GetUserById(ctx, &pb.IdRequest{Id: id})
	if err != nil {
		return nil, err
	}
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
	resp, err := r.client.GetUserByMobile(ctx, &pb.MobileRequest{Mobile: mobile})
	if err != nil {
		return nil, err
	}
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
	resp, err := r.client.CreateUser(ctx, &pb.CreateUserInfo{
		NickName: nickName,
		PassWord: password,
		Mobile:   mobile,
	})
	if err != nil {
		return nil, err
	}
	return &User{
		ID:       resp.Id,
		Mobile:   resp.Mobile,
		Password: resp.PassWord,
		NickName: resp.NickName,
	}, nil
}

var _ UserRepo = (*userRepo)(nil)