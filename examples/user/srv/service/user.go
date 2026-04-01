package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"microg/examples/user/code"
	"microg/examples/user/srv/data"
	merrors "microg/pkg/errors"
)

// UserService 用户业务服务接口
// 定义用户相关的业务操作
type UserService interface {
	List(ctx context.Context, page, pageSize int32) (*data.UserList, error)
	GetByID(ctx context.Context, id int32) (*data.User, error)
	GetByMobile(ctx context.Context, mobile string) (*data.User, error)
	Create(ctx context.Context, nickName, password, mobile string) (*data.User, error)
	Update(ctx context.Context, id int32, nickName, gender string, birthDay int64) error
	CheckPassword(password, encryptedPassword string) bool
}

type userService struct {
	store data.UserStore
}

// NewUserService 创建用户业务服务实例
func NewUserService(store data.UserStore) UserService {
	return &userService{store: store}
}

// List 分页查询用户列表
func (s *userService) List(ctx context.Context, page, pageSize int32) (*data.UserList, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return s.store.List(ctx, page, pageSize)
}

// GetByID 通过ID查询用户
func (s *userService) GetByID(ctx context.Context, id int32) (*data.User, error) {
	return s.store.GetByID(ctx, id)
}

// GetByMobile 通过手机号查询用户
func (s *userService) GetByMobile(ctx context.Context, mobile string) (*data.User, error) {
	return s.store.GetByMobile(ctx, mobile)
}

func (s *userService) Create(ctx context.Context, nickName, password, mobile string) (*data.User, error) {
	_, err := s.store.GetByMobile(ctx, mobile)
	if err == nil {
		return nil, merrors.WithCode(code.ErrUserAlreadyExists, "手机号 %s 已注册", mobile)
	}
	if !merrors.IsCode(err, code.ErrUserNotFound) {
		return nil, err
	}

	user := &data.User{
		Mobile:   mobile,
		Password: encryptPassword(password),
		NickName: nickName,
		Role:     1,
	}
	if err := s.store.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Update 更新用户信息
func (s *userService) Update(ctx context.Context, id int32, nickName, gender string, birthDay int64) error {
	user, err := s.store.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if nickName != "" {
		user.NickName = nickName
	}
	if gender != "" {
		user.Gender = gender
	}
	return s.store.Update(ctx, user)
}

// CheckPassword 校验密码
func (s *userService) CheckPassword(password, encryptedPassword string) bool {
	return encryptPassword(password) == encryptedPassword
}

// encryptPassword 密码加密（MD5）
func encryptPassword(password string) string {
	h := md5.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

var _ UserService = (*userService)(nil)
