package data

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int32          `gorm:"primaryKey;autoIncrement"`
	Mobile    string         `gorm:"uniqueIndex;size:11;not null"`
	Password  string         `gorm:"size:100;not null"`
	NickName  string         `gorm:"size:20"`
	BirthDay  *time.Time     `gorm:"type:datetime"`
	Gender    string         `gorm:"size:10"`
	Role      int32          `gorm:"default:1"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

type UserList struct {
	Total int64
	Items []*User
}

type UserStore interface {
	List(ctx context.Context, page, pageSize int32) (*UserList, error)
	GetByMobile(ctx context.Context, mobile string) (*User, error)
	GetByID(ctx context.Context, id int32) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
}