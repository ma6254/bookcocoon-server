package database

type Book struct {
	ID    uint64 `gorm:"column:id;unique;primaryKey;autoIncrement"` // 书籍ID，主键，自增
	Name  string `gorm:"column:name"`                               // 书籍名称
	Cover string `gorm:"column:cover"`                              // 书籍封面
}
