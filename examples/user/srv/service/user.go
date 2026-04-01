package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/atliliw/microg/examples/user/code"
	"github.com/atliliw/microg/examples/user/srv/data"
	merrors "github.com/atliliw/microg/pkg/errors"
	"github.com/atliliw/microg/pkg/log"
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
	log.DebugC(ctx, "查询用户列表", log.Int32("page", page), log.Int32("pageSize", pageSize))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	result, err := s.store.List(ctx, page, pageSize)
	if err != nil {
		log.ErrorC(ctx, "查询用户列表失败", log.Err(err))
		return nil, err
	}
	log.InfoC(ctx, "查询用户列表成功", log.Int64("total", result.Total))
	return result, nil
}

// GetByID 通过ID查询用户
func (s *userService) GetByID(ctx context.Context, id int32) (*data.User, error) {
	log.DebugC(ctx, "查询用户", log.Int32("id", id))
	user, err := s.store.GetByID(ctx, id)
	if err != nil {
		log.ErrorC(ctx, "查询用户失败", log.Err(err), log.Int32("id", id))
		return nil, err
	}
	log.DebugC(ctx, "查询用户成功", log.Int32("id", id), log.String("nickname", user.NickName))
	return user, nil
}

// GetByMobile 通过手机号查询用户
func (s *userService) GetByMobile(ctx context.Context, mobile string) (*data.User, error) {
	log.DebugC(ctx, "通过手机号查询用户", log.String("mobile", mobile))
	user, err := s.store.GetByMobile(ctx, mobile)
	if err != nil {
		log.ErrorC(ctx, "通过手机号查询用户失败", log.Err(err), log.String("mobile", mobile))
		return nil, err
	}
	log.DebugC(ctx, "通过手机号查询用户成功", log.String("mobile", mobile), log.Int32("id", user.ID))
	return user, nil
}

func (s *userService) Create(ctx context.Context, nickName, password, mobile string) (*data.User, error) {
	log.InfoC(ctx, "创建用户", log.String("nickname", nickName), log.String("mobile", mobile))
	_, err := s.store.GetByMobile(ctx, mobile)
	if err == nil {
		log.WarnC(ctx, "手机号已注册", log.String("mobile", mobile))
		return nil, merrors.WithCode(code.ErrUserAlreadyExists, "手机号 %s 已注册", mobile)
	}
	if !merrors.IsCode(err, code.ErrUserNotFound) {
		log.ErrorC(ctx, "查询手机号失败", log.Err(err), log.String("mobile", mobile))
		return nil, err
	}

	user := &data.User{
		Mobile:   mobile,
		Password: encryptPassword(password),
		NickName: nickName,
		Role:     1,
	}
	if err := s.store.Create(ctx, user); err != nil {
		log.ErrorC(ctx, "创建用户失败", log.Err(err))
		return nil, err
	}
	log.InfoC(ctx, "创建用户成功", log.Int32("id", user.ID), log.String("nickname", nickName))
	return user, nil
}

// Update 更新用户信息
func (s *userService) Update(ctx context.Context, id int32, nickName, gender string, birthDay int64) error {
	log.InfoC(ctx, "更新用户信息", log.Int32("id", id), log.String("nickname", nickName))
	user, err := s.store.GetByID(ctx, id)
	if err != nil {
		log.ErrorC(ctx, "查询用户失败", log.Err(err), log.Int32("id", id))
		return err
	}
	if nickName != "" {
		user.NickName = nickName
	}
	if gender != "" {
		user.Gender = gender
	}
	if err := s.store.Update(ctx, user); err != nil {
		log.ErrorC(ctx, "更新用户失败", log.Err(err), log.Int32("id", id))
		return err
	}
	log.InfoC(ctx, "更新用户成功", log.Int32("id", id))
	return nil
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
