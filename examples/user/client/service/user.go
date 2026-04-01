package service

import (
	"context"

	"microg/examples/user/client/data"
)

// UserService 用户业务服务接口
// 定义客户端的用户相关业务操作
type UserService interface {
	List(ctx context.Context, page, pageSize int32) (*data.UserList, error)
	GetByID(ctx context.Context, id int32) (*data.User, error)
	GetByMobile(ctx context.Context, mobile string) (*data.User, error)
	Create(ctx context.Context, nickName, password, mobile string) (*data.User, error)
}

type userService struct {
	repo data.UserRepo
}

// NewUserService 创建用户业务服务实例
func NewUserService(repo data.UserRepo) UserService {
	return &userService{repo: repo}
}

// List 分页查询用户列表
func (s *userService) List(ctx context.Context, page, pageSize int32) (*data.UserList, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return s.repo.List(ctx, page, pageSize)
}

// GetByID 通过ID查询用户
func (s *userService) GetByID(ctx context.Context, id int32) (*data.User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByMobile 通过手机号查询用户
func (s *userService) GetByMobile(ctx context.Context, mobile string) (*data.User, error) {
	return s.repo.GetByMobile(ctx, mobile)
}

// Create 创建用户
func (s *userService) Create(ctx context.Context, nickName, password, mobile string) (*data.User, error) {
	return s.repo.Create(ctx, nickName, password, mobile)
}

var _ UserService = (*userService)(nil)