package database

import (
	"errors"

)

const (
	WebNovelBookType   = "web_novel"  // 网文小说
	ComicBookType      = "comic"      // 漫画
	PicPackBookType    = "pic_pack"   // 图包
	PublishingBookType = "publishing" // 出版书籍
)

var (
	ErrorBookNotFound = errors.New("book not found")
)

type Book struct {
	ID        uint64 `gorm:"column:id;unique;primaryKey;autoIncrement"` // 书籍ID，主键，自增
	Name      string `gorm:"column:name"`                               // 书籍名称
	Cover     string `gorm:"column:cover"`                              // 书籍封面
	Author    string `gorm:"column:author"`                             // 作者名称
	Type      string `gorm:"column:type"`                               // 书籍类型
	CreatedAt string `gorm:"column:created_at"`                         // 创建时间
	Deleted   bool   `gorm:"column:deleted"`                            // 删除标记，软删除
}

// CreateBook 创建新的书籍信息
func (d *Database) CreateBook(book *Book) error {

	err := d.DB.Create(book).Error
	if err != nil {
		return err
	}
	return nil
}

// GetBookByID 根据书籍ID获取书籍信息
func (d *Database) GetBookByID(book_id uint64) (*Book, error) {
	var (
		err   error
		count int64
		book  Book
	)

	err = d.DB.Model(&Book{}).Where("id = ?", book_id).Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	err = d.DB.Model(&Book{}).Where("id = ?", book_id).First(&book).Error
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// UpdateBook 更新书籍信息
func (d *Database) UpdateBook(book *Book) error {

	var count int64

	// 检查书籍是否存在
	result := d.DB.Model(&Book{}).Where("id = ?", book.ID).Count(&count)
	if result.Error != nil {
		return result.Error
	}
	if count == 0 {
		return ErrorBookNotFound
	}

	// 更新书籍信息
	result = d.DB.Model(&Book{}).Where("id = ?", book.ID).Updates(book)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetAllBooks 获取所有书籍信息
func (d *Database) GetAllBooks() ([]*Book, error) {
	var books []*Book
	result := d.DB.Model(&Book{}).Find(&books)
	if result.Error != nil {
		return nil, result.Error
	}
	return books, nil
}

// CheckBookID 检查书籍ID是否存在
func (d *Database) CheckBookID(book_id uint64) error {
	var count int64
	result := d.DB.Model(&Book{}).Where("id = ?", book_id).Count(&count)
	if result.Error != nil {
		return result.Error
	}
	if count == 0 {
		return ErrorBookNotFound
	}
	return nil
}

// BookDelete 软删除书籍
func (d *Database) BookDelete(book_id uint64) error {
	var count int64
	result := d.DB.Model(&Book{}).Where("id = ?", book_id).Count(&count)
	if result.Error != nil {
		return result.Error
	}
	if count == 0 {
		return ErrorBookNotFound
	}

	// 软删除，设置 Deleted 为 true
	result = d.DB.Model(&Book{}).Where("id = ?", book_id).Update("deleted", true)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
