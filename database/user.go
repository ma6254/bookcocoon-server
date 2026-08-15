package database

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ma6254/bookcocoon-server/validator"
)

type Auth struct {
	ID           uint64 `gorm:"column:id;unique;primaryKey"` // 用户ID，主键，自增
	HashedPasswd string `gorm:"column:hashed_passwd"`        // 用户密码的哈希值
	CreatedAt    string `gorm:"column:created_at"`           // 创建时间
	LoginAt      string `gorm:"column:login_at"`             // 登录时间
	Initnaled    int    `gorm:"column:initnaled"`            // 是否初始化
}

// UserInfoByToken 根据token获取用户信息
func (d *Database) AuthInfoByToken(token string) (*Auth, error) {

	// 查询token对应的用户名
	var t Token
	err := d.DB.Where("token = ?", token).First(&t).Error
	if err != nil {
		return nil, err
	}

	// 查询用户ID对应的用户信息
	var u Auth
	err = d.DB.Where("id = ?", t.ID).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// AuthInfoByName 根据用户名获取用户信息
func (d *Database) AuthInfoByName(user_name string) (*Auth, error) {

	var (
		err   error
		count int64
		user  User
	)

	// 查询用户是否已经存在
	err = d.DB.Model(&User{}).Where("user_name = ?", user_name).Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	// 查询用户是否已经存在
	err = d.DB.Model(&User{}).Where("user_name = ?", user_name).First(&user).Error
	if err != nil {
		return nil, err
	}

	// 查询用户ID对应的用户信息
	var u Auth
	err = d.DB.Model(&Auth{}).Where("id = ?", user.ID).First(&u).Error
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// UserInfoByToken 根据token获取用户信息
func (d *Database) GetAllAuthInfo() ([]*Auth, error) {
	var auths []*Auth
	err := d.DB.Find(&auths).Error
	if err != nil {
		return nil, err
	}
	return auths, nil
}

type User struct {
	ID       uint64 `gorm:"column:id;unique;primaryKey;autoIncrement"` // 用户ID，主键，自增
	UserName string `gorm:"column:user_name;unique"`                   // 用户名，唯一
	NickName string `gorm:"column:nick_name"`                          // 昵称
	Email    string `gorm:"column:email"`                              // 邮箱
}

// UserAuth 检查数据库中的用户名和密码是否正确
func (d *Database) CreateUser(user_name string, hashed_passwd string) (*Auth, error) {

	var (
		err   error
		count int64
	)

	// 查询用户是否已经存在
	err = d.DB.Model(&User{}).Where("user_name = ?", user_name).Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, nil
	}

	now := time.Now().In(d.time_local).Format(time_format)

	auth := &Auth{
		HashedPasswd: hashed_passwd,
		CreatedAt:    now,
		LoginAt:      "",
		Initnaled:    0,
	}

	err = d.DB.Model(&Auth{}).Create(auth).Error
	if err != nil {
		return nil, err
	}

	// 创建对应的用户信息
	user := User{
		ID:       auth.ID,
		UserName: user_name,
		NickName: "",
		Email:    "",
	}

	err = d.DB.Create(&user).Error
	if err != nil {
		return nil, err
	}

	return auth, nil
}

// CreateUserByID 在已知ID的情况下创建用户
func (d *Database) CreateUserByID(user_id uint64, user_name string, hashed_passwd string) (bool, error) {

	var (
		err   error
		count int64
	)

	// 查询用户是否已经存在
	err = d.DB.Model(&Auth{}).Where("id = ?", user_id).Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, fmt.Errorf("user already exists")
	}

	now := time.Now().In(d.time_local).Format(time_format)

	u := &Auth{
		ID:           user_id,
		HashedPasswd: hashed_passwd,
		CreatedAt:    now,
		LoginAt:      "",
		Initnaled:    0,
	}

	err = d.DB.Create(u).Error
	if err != nil {
		return false, err
	}

	user := User{
		ID:       user_id,
		UserName: user_name,
		NickName: "",
		Email:    "admin@example.com",
	}

	err = d.DB.Create(&user).Error
	if err != nil {
		return false, err
	}
	return true, nil
}

// func (d *Database) UserInfo(user_name string) (*User, error) {
// 	var u User
// 	err := d.DB.Where("user_name = ?", user_name).First(&u).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &u, nil
// }

func (d *Database) UpdateUserInfo(user_id uint64, user User) error {

	var (
		err   error
		count int64
	)

	// 查找是否存在这个用户
	err = d.DB.Model(&User{}).Where("id = ?", user_id).Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("user not found")
	}

	// 查找是否有同名用户
	err = d.DB.Model(&User{}).Where("id != ? AND user_name = ?", user_id, user.UserName).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("user name already exists")
	}

	// 查找是否有同邮箱用户
	err = d.DB.Model(&User{}).Where("id != ? AND email = ?", user_id, user.Email).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("email already exists")
	}

	// 更新用户信息
	err = d.DB.Model(&User{}).Where("id = ?", user_id).Updates(user).Error
	if err != nil {
		return err
	}

	return nil
}

// UserInfoByToken 根据token获取用户信息
func (d *Database) UserInfoByToken(token string) (*User, error) {

	// 查询token对应的用户名
	var t Token
	err := d.DB.Where("token = ?", token).First(&t).Error
	if err != nil {
		return nil, err
	}

	// 查询用户名对应的用户信息
	var u User
	err = d.DB.Where("id = ?", t.ID).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// FindUserByID 根据用户ID获取用户信息
func (d *Database) FindUserByID(user_id uint64) (*User, error) {

	var u User
	err := d.DB.Where("id = ?", user_id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetAllUsers 获取所有用户信息
func (d *Database) GetAllUsers() ([]*User, error) {

	var users []*User
	err := d.DB.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserInfoByID 根据用户ID更新用户信息
func (d *Database) UpdateUserInfoByID(user_id uint64, user *User) error {

	var (
		err   error
		count int64
	)

	err = d.DB.Model(&Auth{}).Where("id = ?", user_id).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("user not found")
	}

	err = d.DB.Model(&User{}).Where("id = ?", user_id).Updates(user).Error
	if err != nil {
		return err
	}
	return nil
}

// CheckUserAccount 检查数据库中是否存在指定的用户账号（用户名、邮箱或用户ID）
func (d *Database) CheckUserAccount(account string) (*User, error) {
	var (
		count int64
		user  User
	)

	// 如果是数字，尝试将其解析为用户ID
	user_id, err := strconv.ParseUint(account, 10, 64)
	if err == nil {

		err = d.DB.Model(&User{}).Where("id = ?", user_id).Count(&count).Error
		if err != nil {
			return nil, err
		}

		if count == 0 {
			return nil, nil
		}

		err = d.DB.Model(&User{}).Where("id = ?", user_id).First(&user).Error
		if err != nil {
			return nil, err
		}

		return &user, nil

	}

	err = validator.ValidateUserName(account)
	if err == nil {

		err = d.DB.Model(&User{}).Where("user_name = ?", account).Count(&count).Error
		if err != nil {
			return nil, err
		}

		if count == 0 {
			return nil, nil
		}

		// 如果是合法的用户名，尝试根据用户名查询用户信息
		err = d.DB.Model(&User{}).Where("user_name = ?", account).First(&user).Error
		if err != nil {
			return nil, err
		}

		return &user, nil
	}

	err = validator.ValidateUserEmail(account)
	if err == nil {

		err = d.DB.Model(&User{}).Where("email = ?", account).Count(&count).Error
		if err != nil {
			return nil, err
		}

		if count == 0 {
			return nil, nil
		}

		// 如果是合法的用户名，尝试根据用户名查询用户信息
		err = d.DB.Model(&User{}).Where("email = ?", account).First(&user).Error
		if err != nil {
			return nil, err
		}

		return &user, nil
	}

	// 啥都没找到
	return nil, nil
}

// UserAuth 用户登录，检查数据库中的用户名和密码是否正确
func (d *Database) UserAuth(user_id uint64, hashed_passwd string) (*Auth, error) {

	var (
		err   error
		count int64
		auth  Auth
	)

	// 如果是合法的用户名，尝试根据用户名查询用户信息
	err = d.DB.Model(&Auth{}).Where("id = ? AND hashed_passwd = ?", user_id, hashed_passwd).Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	err = d.DB.Model(&Auth{}).Where("id = ? AND hashed_passwd = ?", user_id, hashed_passwd).First(&auth).Error
	if err != nil {
		return nil, err
	}

	return &auth, nil
}
