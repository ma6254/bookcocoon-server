package database

import (
	"fmt"
	"time"
	_ "time/tzdata"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ma6254/bookcocoon-server/config"
)

const (
	time_format = time.RFC3339 // 时间格式化字符串，使用RFC3339格式
)

type Token struct {
	Token     string `gorm:"column:token;primaryKey;unique"` // token字符串，主键，唯一
	ID        uint64 `gorm:"column:id;unique;primaryKey"`    // 用户ID，主键，自增
	CreatedAt string `gorm:"column:created_at"`              // 创建时间
	UpdatedAt string `gorm:"column:updated_at"`              // 更新时间
	Alive     int    `gorm:"column:alive"`                   // 是否有效
}

type Database struct {
	DB         *gorm.DB
	time_local *time.Location
}

func NewDatabase(cfg *config.DataBase) (*Database, error) {

	var (
		err        error
		db         *gorm.DB
		time_local *time.Location
	)

	time_local, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}

	switch cfg.Driver {
	case config.MYSQL_DRIVER:
		/***********************************************************************
		 * MySQL
		 **********************************************************************/
		mysql_dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.MySQL.User,
			cfg.MySQL.Password,
			cfg.MySQL.Host,
			cfg.MySQL.Port,
			cfg.MySQL.Database,
		)

		db, err = gorm.Open(mysql.Open(mysql_dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}

	case config.SQLITE_DRIVER:
		/***********************************************************************
		 * SQLite
		 **********************************************************************/
		db, err = gorm.Open(sqlite.Open(cfg.SQLite.File), &gorm.Config{})
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	return &Database{
		DB:         db,
		time_local: time_local,
	}, nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// CreateAuthTable 创建认证表
func (d *Database) CreateAuthTable() error {
	err := d.DB.AutoMigrate(&Auth{})
	if err != nil {
		return err
	}
	return nil
}

// CreateUserTable 创建用户表
func (d *Database) CreateUserTable() error {
	err := d.DB.AutoMigrate(&User{})
	if err != nil {
		return err
	}
	return nil
}

// CreateTokenTable 创建token表
func (d *Database) CreateTokenTable() error {
	err := d.DB.AutoMigrate(&Token{})
	if err != nil {
		return err
	}
	return nil
}

// CreateUploadTable 创建upload表
func (d *Database) CreateUploadTable() (bool, error) {
	err := d.DB.AutoMigrate(&UploadFile{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateBookTable 创建book表
func (d *Database) CreateBookTable() (bool, error) {
	err := d.DB.AutoMigrate(&Book{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateReadingRecordTable 创建reading_record表
func (d *Database) CreateReadingRecordTable() (bool, error) {
	err := d.DB.AutoMigrate(&ReadingRecord{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateWebNovelChapterTable 创建web_novel_chapter表
func (d *Database) CreateWebNovelChapterTable() (bool, error) {
	err := d.DB.AutoMigrate(&WebNovelChapter{})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *Database) Install() error {

	var (
		err error
	)

	// 创建认证表
	err = db.CreateAuthTable()
	if err != nil {
		return err
	}

	// 创建用户表
	err = db.CreateUserTable()
	if err != nil {
		return err
	}

	// 创建token表
	err = db.CreateTokenTable()
	if err != nil {
		return err
	}

	// 创建upload表
	_, err = db.CreateUploadTable()
	if err != nil {
		return err
	}

	// 创建book表
	_, err = db.CreateBookTable()
	if err != nil {
		return err
	}

	// 创建web_novel_chapter表
	_, err = db.CreateWebNovelChapterTable()
	if err != nil {
		return err
	}

	// 创建reading_record表
	_, err = db.CreateReadingRecordTable()
	if err != nil {
		return err
	}

	return nil
}

// CreateToken 在数据库中创建一个新的token记录
func (d *Database) CreateToken(token string, user_id uint64) (*Token, error) {
	now := time.Now().In(d.time_local).Format(time_format)
	t := &Token{
		Token:     token,
		ID:        user_id,
		CreatedAt: now,
		UpdatedAt: now,
		Alive:     1,
	}
	err := d.DB.Create(t).Error
	if err != nil {
		return nil, err
	}
	return t, nil
}

// CheckToken 检查数据库中的token是否有效
func (d *Database) CheckToken(token string, user_id uint64) (bool, error) {
	var count int64
	err := d.DB.Model(&Token{}).Where("token = ? AND id = ?", token, user_id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *Database) GetUserAllAliveToken(user_id uint64) ([]*Token, error) {
	var tokens []*Token
	err := d.DB.Model(&Token{}).Where("id = ? AND alive = 1", user_id).Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}
