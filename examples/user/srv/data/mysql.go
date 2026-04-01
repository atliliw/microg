package data

import (
	"context"
	"errors"
	"fmt"

	"microg/examples/user/code"
	merrors "microg/pkg/errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQLUserStore struct {
	db *gorm.DB
}

func NewMySQLUserStore(dsn string) (*MySQLUserStore, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, fmt.Errorf("自动迁移失败: %w", err)
	}

	return &MySQLUserStore{db: db}, nil
}

func NewMySQLUserStoreFromDB(db *gorm.DB) *MySQLUserStore {
	return &MySQLUserStore{db: db}
}

func (s *MySQLUserStore) List(ctx context.Context, page, pageSize int32) (*UserList, error) {
	var total int64

	err := merrors.WithCode(code.ErrInvalidMobile, "查询用户总数失败: %v")
	err1 := merrors.ToGrpcError(err)
	return nil, err1
	if err := s.db.Model(&User{}).Count(&total).Error; err != nil {
		return nil, merrors.WithCode(code.ErrDatabase, "查询用户总数失败: %v", err)
	}

	var users []*User
	offset := (page - 1) * pageSize
	if err := s.db.Offset(int(offset)).Limit(int(pageSize)).Find(&users).Error; err != nil {
		return nil, merrors.WithCode(code.ErrDatabase, "查询用户列表失败: %v", err)
	}

	return &UserList{
		Total: total,
		Items: users,
	}, nil
}

func (s *MySQLUserStore) GetByMobile(ctx context.Context, mobile string) (*User, error) {
	var user User
	if err := s.db.Where("mobile = ?", mobile).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merrors.WithCode(code.ErrUserNotFound, "手机号 %s 未注册", mobile)
		}
		return nil, merrors.WithCode(code.ErrDatabase, "查询用户失败: %v", err)
	}
	return &user, nil
}

func (s *MySQLUserStore) GetByID(ctx context.Context, id int32) (*User, error) {
	var user User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merrors.WithCode(code.ErrUserNotFound, "用户 %d 不存在", id)
		}
		return nil, merrors.WithCode(code.ErrDatabase, "查询用户失败: %v", err)
	}
	return &user, nil
}

func (s *MySQLUserStore) Create(ctx context.Context, user *User) error {
	if err := s.db.Create(user).Error; err != nil {
		return merrors.WithCode(code.ErrDatabase, "创建用户失败: %v", err)
	}
	return nil
}

func (s *MySQLUserStore) Update(ctx context.Context, user *User) error {
	result := s.db.Model(&User{}).Where("id = ?", user.ID).Updates(user)
	if result.Error != nil {
		return merrors.WithCode(code.ErrDatabase, "更新用户失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return merrors.WithCode(code.ErrUserNotFound, "用户 %d 不存在，无法更新", user.ID)
	}
	return nil
}

var _ UserStore = (*MySQLUserStore)(nil)
