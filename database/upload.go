package database

import (
	"time"

	"gorm.io/gorm"
)

// Upload 上传文件信息
type UploadFile struct {
	FileID      int64  `gorm:"column:file_id;unique;primaryKey"` // 文件ID，主键
	Hash        string `gorm:"column:hash;unique"`               // 文件哈希，主键
	Name        string `gorm:"column:name"`                      // 文件名
	Size        uint64 `gorm:"column:size"`                      // 文件大小
	Path        string `gorm:"column:path"`                      // 文件存储路径
	CreatedAt   string `gorm:"column:created_at"`                // 创建日期
	UploadedAt  string `gorm:"column:uploaded_at"`               // 上传日期
	CreatedByID uint64 `gorm:"column:created_by_id"`             // 上传者ID，外键关联到用户表
}

// FindUploadByHash 根据文件哈希查找上传文件信息
func (d *Database) CreateUpload(file_id int64, hash string, name string, size uint64, path string, user_id uint64) (*UploadFile, error) {

	// 检查是否已存在相同哈希的文件
	existingUpload, err := d.FindUploadByHash(hash)
	if err != nil {
		return nil, err
	}
	if existingUpload != nil {
		return existingUpload, nil // 文件已存在，返回现有文件信息
	}

	// 检查用户是否存在
	existingUser, err := d.FindUserByID(user_id)
	if err != nil || existingUser == nil {
		return nil, err // 用户不存在，返回错误
	}

	upload := &UploadFile{
		FileID:      file_id,
		Hash:        hash,
		Name:        name,
		Size:        size,
		Path:        path,
		CreatedAt:   time.Now().Format(time_format),
		UploadedAt:  "",
		CreatedByID: user_id,
	}

	result := d.DB.Create(upload)
	if result.Error != nil {
		return nil, result.Error
	}

	return upload, nil
}

func (d *Database) UpdateUploadTime(file_id int64) (*UploadFile, error) {
	upload, err := d.FindUploadByID(file_id)
	if err != nil {
		return nil, err
	}
	if upload == nil {
		return nil, nil // 文件信息不存在，返回nil
	}

	upload.UploadedAt = time.Now().Format(time_format)
	result := d.DB.Save(upload)
	if result.Error != nil {
		return nil, result.Error
	}
	return upload, nil
}

func (d *Database) UpdateUploadSize(file_id int64, size uint64) (*UploadFile, error) {
	upload, err := d.FindUploadByID(file_id)
	if err != nil {
		return nil, err
	}
	if upload == nil {
		return nil, nil // 文件信息不存在，返回nil
	}

	upload.Size = size
	result := d.DB.Save(upload)
	if result.Error != nil {
		return nil, result.Error
	}
	return upload, nil
}

// FindUploadByHash 根据文件哈希查找上传文件信息
func (d *Database) FindUploadByHash(hash string) (*UploadFile, error) {

	var (
		count  int64
		result *gorm.DB
	)

	// 是否存在文件信息
	result = d.DB.Model(&UploadFile{}).Where("hash = ?", hash).Count(&count)
	if result.Error != nil {
		return nil, result.Error
	}
	if count == 0 {
		return nil, nil // 文件信息不存在，返回nil
	}

	// 获取上传文件信息
	var upload UploadFile
	result = d.DB.Where("hash = ?", hash).First(&upload)
	if result.Error != nil {
		return nil, result.Error
	}
	return &upload, nil
}

// FindUploadByID 根据文件ID查找上传文件信息
func (d *Database) FindUploadByID(file_id int64) (*UploadFile, error) {

	var (
		count  int64
		result *gorm.DB
	)

	// 是否存在文件信息
	result = d.DB.Model(&UploadFile{}).Where("file_id = ?", file_id).Count(&count)
	if result.Error != nil {
		return nil, result.Error
	}
	if count == 0 {
		return nil, nil // 文件信息不存在，返回nil
	}

	// 获取上传文件信息
	var upload UploadFile
	result = d.DB.Where("file_id = ?", file_id).First(&upload)
	if result.Error != nil {
		return nil, result.Error
	}
	return &upload, nil
}

// GetAllUploads 获取所有上传文件信息
func (d *Database) GetAllUploads() ([]UploadFile, error) {
	var uploads []UploadFile
	result := d.DB.Find(&uploads)
	if result.Error != nil {
		return nil, result.Error
	}
	return uploads, nil
}
