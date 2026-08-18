package database

// 网文章节信息
type WebNovelChapter struct {
	ID     uint64 `gorm:"column:id;unique;primaryKey;autoIncrement"` // 章节ID，唯一主键自增长
	BookID uint64 `gorm:"column:book_id"`                            // 关联的书籍ID
	Index  uint64 `gorm:"column:index"`                              // 这本小说里的章节索引
	Name   string `gorm:"column:name"`                               // 章节名称
}

// CreateWebNovelChapter 创建新的网文章节信息
func (d *Database) CreateWebNovelChapter(chapter *WebNovelChapter) error {

	err := d.DB.Create(chapter).Error
	if err != nil {
		return err
	}
	return nil
}

// GetWebNovelChaptersByBookID 获取指定书籍的所有章节
func (d *Database) GetWebNovelChaptersByBookID(bookID uint64) ([]*WebNovelChapter, error) {

	var chapters []*WebNovelChapter

	// index是SQL的关键字，所以需要加上反引号
	err := d.DB.Model(&WebNovelChapter{}).Where("book_id = ?", bookID).Order("`index` ASC").Find(&chapters).Error
	if err != nil {
		return nil, err
	}
	return chapters, nil
}

// GetWebNovelChapters 获取指定书籍的指定章节
func (d *Database) GetWebNovelChaptersBookIndex(bookID uint64, index uint64) (*WebNovelChapter, error) {

	var count int64
	err := d.DB.Model(&WebNovelChapter{}).Where("book_id = ? AND `index` = ?", bookID, index).Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	var chapters WebNovelChapter
	err = d.DB.Model(&WebNovelChapter{}).Where("book_id = ? AND `index` = ?", bookID, index).First(&chapters).Error
	if err != nil {
		return nil, err
	}

	return &chapters, nil
}

// GetWebNovelChaptersByBookID 获取指定书籍的所有章节
func (d *Database) DeleteBookChaptersByBookID(bookID uint64) error {

	// 删除章节
	err := d.DB.Where("book_id = ?", bookID).Delete(&WebNovelChapter{}).Error
	if err != nil {
		return err
	}

	return nil
}
