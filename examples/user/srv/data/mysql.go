package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/atliliw/microg/examples/user/code"
	merrors "github.com/atliliw/microg/pkg/errors"
	"github.com/atliliw/microg/pkg/log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQLUserStore struct {
	db *gorm.DB
}

func NewMySQLUserStore(database string, dsn string, noDatabaseDSN string) (*MySQLUserStore, error) {
	if err := createDatabaseIfNotExists(database, noDatabaseDSN); err != nil {
		return nil, fmt.Errorf("创建数据库失败: %w", err)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, fmt.Errorf("自动迁移失败: %w", err)
	}

	if err := seedInitialData(db); err != nil {
		log.Warn("初始化数据失败", log.Err(err))
	}

	return &MySQLUserStore{db: db}, nil
}

func createDatabaseIfNotExists(database string, noDatabaseDSN string) error {
	db, err := gorm.Open(mysql.Open(noDatabaseDSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接 MySQL 服务器失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}
	defer sqlDB.Close()

	var exists int
	query := fmt.Sprintf("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '%s'", database)
	if err := sqlDB.QueryRow(query).Scan(&exists); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("检查数据库是否存在失败: %w", err)
	}

	if exists == 0 {
		createSQL := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", database)
		if _, err := sqlDB.Exec(createSQL); err != nil {
			return fmt.Errorf("创建数据库失败: %w", err)
		}
		log.Info("数据库创建成功", log.String("database", database))
	}

	return nil
}

func seedInitialData(db *gorm.DB) error {
	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("检查用户数量失败: %w", err)
	}

	if count > 0 {
		return nil
	}

	users := []*User{
		{
			Mobile:   "13800138001",
			Password: "password123",
			NickName: "张三",
			Gender:   "male",
			Role:     1,
		},
		{
			Mobile:   "13800138002",
			Password: "password123",
			NickName: "李四",
			Gender:   "male",
			Role:     1,
		},
		{
			Mobile:   "13800138003",
			Password: "password123",
			NickName: "王五",
			Gender:   "female",
			Role:     1,
		},
		{
			Mobile:   "13800138004",
			Password: "password123",
			NickName: "管理员",
			Gender:   "male",
			Role:     2,
		},
	}

	if err := db.Create(&users).Error; err != nil {
		return fmt.Errorf("创建初始用户失败: %w", err)
	}

	log.Info("初始化数据成功", log.Int("count", len(users)))
	return nil
}

func NewMySQLUserStoreFromDB(db *gorm.DB) *MySQLUserStore {
	return &MySQLUserStore{db: db}
}

func (s *MySQLUserStore) List(ctx context.Context, page, pageSize int32) (*UserList, error) {
	var total int64

	//err := merrors.WithCode(code.ErrInvalidMobile, "查询用户总数失败: %v")
	//err1 := merrors.ToGrpcError(err)
	//return nil, err1
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
