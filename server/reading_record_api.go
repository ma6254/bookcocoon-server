package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ma6254/bookcocoon-server/database"
)

type ReadingRecordCreateForm struct {
	BookID uint64 `json:"book_id"` // 书籍ID
}

type ReadingRecordUpdateForm struct {
	BookIndex uint64 `json:"book_index"` // 关联的书籍章节索引
}

type ReadingRecord struct {
	UserID    uint64 `json:"user_id"`    // 关联的用户ID
	BookID    uint64 `json:"book_id"`    // 关联的书籍ID
	BookIndex uint64 `json:"book_index"` // 关联的书籍章节索引
	CreatedAt string `json:"created_at"` // 创建时间
	UpdatedAt string `json:"updated_at"` // 更新时间
}

// NewBookInfoByDB 根据数据库中的Book对象创建API响应的Book对象
func NewReadingRecordByDB(db_reading_record *database.ReadingRecord) *ReadingRecord {

	return &ReadingRecord{
		UserID:    db_reading_record.UserID,
		BookID:    db_reading_record.BookID,
		BookIndex: db_reading_record.BookIndex,
		CreatedAt: db_reading_record.CreatedAt,
		UpdatedAt: db_reading_record.UpdatedAt,
	}
}

// http_reading_record_create_api_handler 创建阅读记录
// @Summary      创建阅读记录
// @Description  创建阅读记录。
// @Tags         阅读记录
// @Accept       json
// @Produce      json
// @Param        body  body      ReadingRecordCreateForm  true  "阅读记录请求"
// @Success      200  {object}  ReadingRecord
// @Router       /reading_record/create [post]
// @Security     Bearer
func http_reading_record_create_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err error
		req = ReadingRecordCreateForm{}
	)

	// 解码json请求体
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		HttpErrorJsonDecode(w, err)
		return
	}

	// 书籍ID
	if req.BookID == 0 {
		http.Error(w, "book_id is required", http.StatusBadRequest)
		return
	}

	// 检查书籍ID书否存在
	book, err := s.DB.GetBookByID(req.BookID)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}

	if book == nil {
		HttpErrorBookNotFound(w)
		return
	}

	// 创建阅读记录
	readingRecord, err := s.DB.CreateReadingRecord(session.UserID, req.BookID)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}
	if readingRecord == nil {
		http.Error(w, "reading record already exists", http.StatusConflict)
		return
	}

	// 返回创建的阅读记录
	s.WriteJsonSuccessResponse(w, NewReadingRecordByDB(readingRecord))
}

// http_reading_record_get_api_handler 获取本用户指定书籍阅读记录
// @Summary      获取阅读记录
// @Description  获取阅读记录。
// @Tags         阅读记录
// @Accept       json
// @Produce      json
// @Param        book_id  path      uint64  true  "书籍ID"
// @Success      200  {object}  ReadingRecord
// @Router       /reading_record/{book_id} [get]
// @Security     Bearer
func http_reading_record_get_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err         error
		book_id_str = r.PathValue("book_id")
	)

	book_id, err := strconv.ParseUint(book_id_str, 10, 64)
	if err != nil {
		http.Error(w, "invalid book_id", http.StatusBadRequest)
		return
	}

	// 获取阅读记录
	readingRecord, err := s.DB.GetReadingRecord(session.UserID, book_id)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}
	if readingRecord == nil {
		http.Error(w, "reading record not found", http.StatusNotFound)
		return
	}

	// 返回阅读记录
	s.WriteJsonSuccessResponse(w, NewReadingRecordByDB(readingRecord))
}

// http_reading_record_get_api_handler 获取本用户所有书籍阅读记录
// @Summary      获取阅读记录
// @Description  获取阅读记录。
// @Tags         阅读记录
// @Accept       json
// @Produce      json
// @Success      200  {array}  ReadingRecord
// @Router       /reading_record [get]
// @Security     Bearer
func http_reading_record_list_by_user_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {
	var readingRecords, err = s.DB.GetReadingRecordListByUser(session.UserID)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}

	record_list := []*ReadingRecord{}

	for _, readingRecord := range readingRecords {
		record_list = append(record_list, NewReadingRecordByDB(readingRecord))
	}

	s.WriteJsonSuccessResponse(w, record_list)
}

// http_reading_record_update_api_handler 更新阅读记录
// @Summary      更新阅读记录
// @Description  更新阅读记录。
// @Tags         阅读记录
// @Accept       json
// @Produce      json
// @Param        book_id  path      uint64  true  "书籍ID"
// @Success      200  {object}  ReadingRecord
// @Router       /reading_record/{book_id} [post]
// @Security     Bearer
func http_reading_record_update_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err         error
		book_id_str = r.PathValue("book_id")
		from        ReadingRecordUpdateForm
	)

	if err := json.NewDecoder(r.Body).Decode(&from); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	book_id, err := strconv.ParseUint(book_id_str, 10, 64)
	if err != nil {
		http.Error(w, "invalid book_id", http.StatusBadRequest)
		return
	}

	// 获取阅读记录
	readingRecord, err := s.DB.GetReadingRecord(session.UserID, book_id)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}
	if readingRecord == nil {
		http.Error(w, "reading record not found", http.StatusNotFound)
		return
	}

	// 更新阅读记录
	readingRecord, err = s.DB.UpdateReadingRecord(session.UserID, book_id, from.BookIndex)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}
	if readingRecord == nil {
		http.Error(w, "reading record not found", http.StatusNotFound)
		return
	}

	// 返回更新后的阅读记录
	s.WriteJsonSuccessResponse(w, NewReadingRecordByDB(readingRecord))
}
