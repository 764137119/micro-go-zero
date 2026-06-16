package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

/*
CREATE TABLE `user_info` (
    `user_id`    BIGINT NOT NULL AUTO_INCREMENT COMMENT '用户ID（主键）',
    `mobile`     VARCHAR(20) NOT NULL COMMENT '手机号',
    `nick_name`  VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户昵称',
    `sex`        INT NOT NULL DEFAULT 0 COMMENT '性别：0-未知, 1-男, 2-女',
    `avatar`     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像URL',
    `email`      VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
    `password`   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '密码（加密存储）',
    `created_at` BIGINT NOT NULL COMMENT '创建时间（毫秒时间戳）',
    `updated_at` BIGINT NOT NULL COMMENT '更新时间（毫秒时间戳）',
    PRIMARY KEY (`user_id`),
    UNIQUE KEY `idx_mobile` (`mobile`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
*/

// User 用户表模型
type User struct {
	UserId    int64  `gorm:"primaryKey;autoIncrement;column:user_id"`               // 用户ID（主键）
	Mobile    string `gorm:"column:mobile;type:varchar(20);not null;uniqueIndex"`   // 手机号（唯一索引）
	NickName  string `gorm:"column:nick_name;type:varchar(64);not null;default:''"` // 用户昵称
	Sex       int32  `gorm:"column:sex;type:int;not null;default:0"`                // 性别：0-未知, 1-男, 2-女
	Avatar    string `gorm:"column:avatar;type:varchar(255);not null;default:''"`   // 头像URL
	Email     string `gorm:"column:email;type:varchar(128);not null;default:''"`    // 邮箱
	Password  string `gorm:"column:password;type:varchar(128);not null;default:''"` // 密码（加密存储）
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime;milli"`                // 创建时间（毫秒时间戳）
	UpdatedAt int64  `gorm:"column:updated_at;autoUpdateTime;milli"`                // 更新时间（毫秒时间戳）
}

// TableName 自定义表名
func (User) TableName() string {
	return "user_info"
}

// UserRepo 用户仓储层
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓储
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create 创建用户
func (r *UserRepo) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByUserId 根据用户ID查询
func (r *UserRepo) FindByUserId(ctx context.Context, userId int64) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("user_id = ?", userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByMobile 根据手机号查询
func (r *UserRepo) FindByMobile(ctx context.Context, mobile string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("mobile = ?", mobile).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *UserRepo) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Model(&User{}).
		Where("user_id = ?", user.UserId).
		Updates(map[string]interface{}{
			"nick_name":  user.NickName,
			"sex":        user.Sex,
			"avatar":     user.Avatar,
			"email":      user.Email,
			"password":   user.Password,
			"updated_at": time.Now().UnixMilli(),
		}).Error
}

// FindList 分页查询用户列表
func (r *UserRepo) FindList(ctx context.Context, page, pageSize int32, mobile, nickName string) ([]*User, int64, error) {
	var list []*User
	var total int64
	query := r.db.WithContext(ctx).Model(&User{})

	if mobile != "" {
		query = query.Where("mobile LIKE ?", "%"+mobile+"%")
	}
	if nickName != "" {
		query = query.Where("nick_name LIKE ?", "%"+nickName+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Order("user_id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
