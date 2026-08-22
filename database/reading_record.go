package database

import "time"

type ReadingRecord struct {
	ID        uint64 `gorm:"column:id;unique;primaryKey;autoIncrement"` // 章节ID，唯一主键自增长
	UserID    uint64 `gorm:"column:user_id"`                            // 关联的用户ID
	BookID    uint64 `gorm:"column:book_id"`                            // 关联的书籍ID
	BookIndex uint64 `gorm:"column:book_index"`                         // 关联的书籍章节索引
	CreatedAt string `gorm:"column:created_at"`                         // 创建时间
	UpdatedAt string `gorm:"column:updated_at"`                         // 更新时间
}

// CreateReadingRecord 创建新的阅读记录
func (d *Database) CreateReadingRecord(user_id uint64, book_id uint64) (*ReadingRecord, error) {

	var (
		err   error
		count int64
	)

	// 查询是否已经存在该用户和书籍的阅读记录
	err = d.DB.Model(&ReadingRecord{}).Where("user_id = ? AND book_id = ?", user_id, book_id).Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, nil
	}

	now := time.Now().In(d.time_local).Format(time_format)

	readingRecord := &ReadingRecord{
		UserID:    user_id,
		BookID:    book_id,
		CreatedAt: now,
	}

	err = d.DB.Model(&ReadingRecord{}).Create(readingRecord).Error
	if err != nil {
		return nil, err
	}

	return readingRecord, nil
}

// UpdateReadingRecord 更新阅读记录
func (d *Database) UpdateReadingRecord(user_id uint64, book_id uint64, book_index uint64) (*ReadingRecord, error) {
	now := time.Now().In(d.time_local).Format(time_format)

	readingRecord := &ReadingRecord{
		UserID:    user_id,
		BookID:    book_id,
		BookIndex: book_index,
		UpdatedAt: now,
	}

	err := d.DB. //
			Model(&ReadingRecord{}).                                //
			Where("user_id = ? AND book_id = ?", user_id, book_id). //
			Select("book_index", "updated_at").                     //
			Updates(readingRecord).Error
	if err != nil {
		return nil, err
	}

	err = d.DB.Model(&ReadingRecord{}).Where("user_id = ? AND book_id = ?", user_id, book_id).First(readingRecord).Error
	if err != nil {
		return nil, err
	}

	return readingRecord, nil

}

// GetReadingRecord 获取阅读记录
func (d *Database) GetReadingRecord(user_id uint64, book_id uint64) (*ReadingRecord, error) {

	var count int64

	err := d.DB.Model(&ReadingRecord{}).Where("user_id = ? AND book_id = ?", user_id, book_id).Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	readingRecord := &ReadingRecord{}
	err = d.DB.Model(&ReadingRecord{}).Where("user_id = ? AND book_id = ?", user_id, book_id).First(readingRecord).Error
	if err != nil {
		return nil, err
	}

	return readingRecord, nil

}

// GetReadingRecordListByUser 获取用户的所有阅读记录
func (d *Database) GetReadingRecordListByUser(user_id uint64) ([]*ReadingRecord, error) {
	var readingRecords []*ReadingRecord
	err := d.DB.Model(&ReadingRecord{}).Where("user_id = ?", user_id).Find(&readingRecords).Error
	if err != nil {
		return nil, err
	}
	return readingRecords, nil
}
	