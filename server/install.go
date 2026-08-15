package server

import (
	"os"

	"github.com/ma6254/bookcocoon-server/database"
)

const (
	// 管理员用户ID
	// 第一个用户，其他用户ID从10001开始
	admin_user_id = 10000

	// 管理员用户默认密码
	admin_user_default_passwd = "1234567890"
)

// IsInstalled 检查服务器是否已经安装
func (s *Server) IsInstalled() bool {
	switch s.Config.DataBase.Driver {
	case "sqlite":
		if _, err := os.Stat(s.Config.DataBase.SQLite.File); os.IsNotExist(err) {
			return false
		}
	case "mysql":
		// 可以在这里添加 MySQL 的安装检查逻辑
	}

	return true
}

// Install 安装服务器，包括创建数据库文件和初始化数据库表
func (s *Server) Install() error {

	// 如果服务器已经安装，则不需要再次安装
	if s.is_installed == true {
		return nil
	}

	err := s.DB.Install()
	if err != nil {
		return err
	}

	user_name := "admin"

	hashed_passwd := s.Salt(admin_user_id, admin_user_default_passwd)

	// 创建默认用户
	_, err = s.DB.CreateUserByID(admin_user_id, user_name, hashed_passwd)
	if err != nil {
		return err
	}

	user := database.User{
		ID:       admin_user_id,
		UserName: user_name,
		NickName: "管理员",
		Email:    "admin@example.com",
	}

	err = s.DB.UpdateUserInfo(admin_user_id, user)
	if err != nil {
		return err
	}

	s.is_installed = true
	return nil
}
