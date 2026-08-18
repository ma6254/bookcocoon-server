package server

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ma6254/bookcocoon-server/database"
	"github.com/ma6254/bookcocoon-server/web_novel_book"
	"gorm.io/gorm"
)

type CreateBookForm struct {
	Name   string `json:"name"`   // 书籍名称
	Cover  string `json:"cover"`  // 书籍封面
	Author string `json:"author"` // 作者名称
	Type   string `json:"type"`   // 书籍类型
}

type UpdateBookForm struct {
	Name   string `json:"name"`   // 书籍名称
	Cover  string `json:"cover"`  // 书籍封面
	Author string `json:"author"` // 作者名称
	Type   string `json:"type"`   // 书籍类型
}

type Book struct {
	ID        uint64 `json:"id"`         // 书籍ID
	Name      string `json:"name"`       // 书籍名称
	Cover     string `json:"cover"`      // 书籍封面
	Author    string `json:"author"`     // 作者名称
	Type      string `json:"type"`       // 书籍类型
	CreatedAt string `json:"created_at"` // 创建时间
}

type Chapter struct {
	Index  uint64 `json:"index"`   // 章节索引
	BookID uint64 `json:"book_id"` // 书籍ID
	Title  string `json:"title"`   // 章节标题
}

// NewBookInfoByDB 根据数据库中的Book对象创建API响应的Book对象
func NewBookInfoByDB(db_book *database.Book) *Book {

	return &Book{
		ID:        db_book.ID,
		Name:      db_book.Name,
		Cover:     db_book.Cover,
		Author:    db_book.Author,
		Type:      db_book.Type,
		CreatedAt: db_book.CreatedAt,
	}
}

// book_cover_ext_check 检查封面文件扩展名是否合法
func book_cover_ext_check(cover_file_name string) bool {
	ext := strings.ToLower(path.Ext(cover_file_name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return true
	default:
		return false
	}
}

// book_raw_file_ext_check 检查原始文件扩展名是否合法
func book_raw_file_ext_check(book_type string, raw_file_name string) bool {

	if book_type == database.WebNovelBookType {
		ext := strings.ToLower(path.Ext(raw_file_name))

		if ext == ".txt" {
			return true
		}
		return false
	}

	return false
}

// http_book_create_api_handler 创建书籍
// @Summary      创建书籍
// @Description  创建新的书籍信息。
// @Tags         书籍
// @Accept       json
// @Produce      json
// @Param        body  body      CreateBookForm  true  "创建书籍表单"
// @Success      200   {object}  Book
// @Router       /book/create [post]
// @Security     Bearer
func http_book_create_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err     error
		book_id uint64
	)

	form := &CreateBookForm{}

	err = json.NewDecoder(r.Body).Decode(&form)
	if err != nil {
		http.Error(w, "json decode error: "+err.Error(), http.StatusBadRequest)
		return
	}

	db_book := &database.Book{
		Name:      form.Name,
		Cover:     form.Cover,
		Author:    form.Author,
		Type:      form.Type,
		CreatedAt: time.Now().Format(time_format),
	}

	err = s.DB.CreateBook(db_book)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	book_id = db_book.ID

	book_path := path.Join(books_path, strconv.FormatUint(book_id, 10))
	os.MkdirAll(book_path, os.ModePerm)

	s.WriteJsonSuccessResponse(w, NewBookInfoByDB(db_book))
}

// http_book_list_api_handler 书籍列表
// @Summary      书籍列表
// @Description  获取所有书籍信息。
// @Tags         书籍
// @Accept       json
// @Produce      json
// @Success      200  {object} []Book
// @Router       /book/list [get]
// @Security     Bearer
func http_book_list_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	books, err := s.DB.GetAllBooks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []*Book

	for _, book := range books {
		response = append(response, NewBookInfoByDB(book))
	}

	s.WriteJsonSuccessResponse(w, response)
}

// http_book_update_api_handler 更新书籍
// @Summary      更新书籍
// @Description  更新已有的书籍信息。
// @Tags         书籍
// @Accept       json
// @Produce      json
// @Param        book_id  path      uint64          true  "书籍ID"
// @Param        body  body      UpdateBookForm  true  "更新书籍表单"
// @Success      200   {object}  Book
// @Router       /book/update/{book_id} [post]
// @Security     Bearer
func http_book_update_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err         error
		book_id_str = r.PathValue("book_id")
	)

	book_id, err := strconv.ParseUint(book_id_str, 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// 检查书籍是否存在
	err = s.DB.CheckBookID(book_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	form := &UpdateBookForm{}
	err = json.NewDecoder(r.Body).Decode(&form)
	if err != nil {
		http.Error(w, "json decode error: "+err.Error(), http.StatusBadRequest)
		return
	}

	db_book := &database.Book{
		ID:     book_id,
		Name:   form.Name,
		Cover:  form.Cover,
		Author: form.Author,
		Type:   form.Type,
	}

	err = s.DB.UpdateBook(db_book)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.WriteJsonSuccessResponse(w, NewBookInfoByDB(db_book))
}

// http_book_update_cover_api_handler 更新书籍封面
// @Summary      更新书籍封面
// @Description  更新已有的书籍封面。
// @Tags         书籍
// @Accept       multipart/form-data
// @Produce      json
// @Param        book_id  path      uint64          true  "书籍ID"
// @Param        file     formData  file    true  "上传的文件"
// @Success      200
// @Router       /book/cover/{book_id} [post]
// @Security     Bearer
func http_book_update_cover_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err         error
		book_id_str = r.PathValue("book_id")
	)

	book_id, err := strconv.ParseUint(book_id_str, 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// 检查书籍是否存在
	err = s.DB.CheckBookID(book_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var coverReader io.Reader
	var file_header *multipart.FileHeader

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		file, fileHeader, fileErr := r.FormFile("file")
		if fileErr != nil {
			http.Error(w, "Invalid multipart file data", http.StatusBadRequest)
			return
		}
		defer file.Close()
		coverReader = file
		file_header = fileHeader
	} else {
		http.Error(w, "Content-Type must be multipart/form-data", http.StatusBadRequest)
		return
	}

	var upload_file_name = file_header.Filename
	var upload_file_ext = strings.ToLower(path.Ext(upload_file_name))

	// 检查封面文件扩展名是否合法
	if book_cover_ext_check(upload_file_name) == false {
		http.Error(w, "Invalid cover file extension", http.StatusBadRequest)
		return
	}

	// 读取封面数据
	cover_data, err := io.ReadAll(coverReader)
	if err != nil {
		http.Error(w, "read cover data error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 清除原有封面文件
	book_path := path.Join(books_path, book_id_str)
	files, err := os.ReadDir(book_path)
	if err != nil {
		http.Error(w, "read book directory error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, f := range files {

		// 这里只删除以 "cover." 开头的文件，并且扩展名合法的文件
		if strings.HasPrefix(f.Name(), "cover.") && book_cover_ext_check(f.Name()) {
			os.Remove(path.Join(book_path, f.Name()))
		}
	}

	// 保存封面文件到书籍目录
	err = os.WriteFile(path.Join(books_path, book_id_str, "cover"+upload_file_ext), cover_data, 0644)
	if err != nil {
		http.Error(w, "write cover file error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.WriteSuccessResponse(w)
}

// http_book_get_cover_api_handler 获取书籍封面
// @Summary      获取书籍封面
// @Description  获取已有的书籍封面。
// @Tags         书籍
// @Accept       json
// @Produce      json
// @Param        book_id  path      uint64          true  "书籍ID"
// @Success      200
// @Router       /book/cover/{book_id} [get]
// @Security     Bearer
func http_book_get_cover_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err         error
		book_id_str = r.PathValue("book_id")
	)

	book_id, err := strconv.ParseUint(book_id_str, 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// 检查书籍是否存在
	err = s.DB.CheckBookID(book_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 查找书籍封面文件
	book_cover_path := ""
	book_path := path.Join(books_path, book_id_str)
	files, err := os.ReadDir(book_path)
	if err != nil {
		http.Error(w, "read book directory error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), "cover.") && book_cover_ext_check(f.Name()) {
			book_cover_path = path.Join(book_path, f.Name())
			break
		}
	}

	if book_cover_path == "" {
		http.Error(w, "cover file not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, book_cover_path)
}

// http_book_update_raw_api_handler 更新书籍原始文件
// @Summary      更新书籍原始文件
// @Description  更新已有的书籍原始文件。
// @Tags         书籍
// @Accept       multipart/form-data
// @Produce      json
// @Param        book_id  path      uint64          true  "书籍ID"
// @Param        file     formData  file    true  "上传的文件"
// @Success      200   {object}  Book
// @Router       /book/raw/{book_id} [post]
// @Security     Bearer
func http_book_update_raw_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err         error
		book_id_str = r.PathValue("book_id")
	)

	book_id, err := strconv.ParseUint(book_id_str, 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// 检查书籍是否存在
	err = s.DB.CheckBookID(book_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 检查Content-Type是否为multipart/form-data
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		http.Error(w, "Content-Type must be multipart/form-data", http.StatusBadRequest)
		return
	}

	// 检查书籍类型
	db_book, err := s.DB.GetBookByID(book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 清除原有原始文件
	book_path := path.Join(books_path, book_id_str)
	files, err := os.ReadDir(book_path)
	if err != nil {
		http.Error(w, "read book directory error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, f := range files {

		// 这里只删除以 "raw." 开头的文件，并且扩展名合法的文件
		if strings.HasPrefix(f.Name(), "raw.") && book_raw_file_ext_check(db_book.Type, f.Name()) {
			os.Remove(path.Join(book_path, f.Name()))
		}
	}

	// 保存新的原始文件
	file, file_header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "get form file error: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 检查原始文件扩展名是否合法
	ok := book_raw_file_ext_check(db_book.Type, file_header.Filename)
	if ok == false {
		http.Error(w, "Invalid raw file extension for book type", http.StatusBadRequest)
		return
	}

	raw_file_path := path.Join(book_path, "raw"+path.Ext(file_header.Filename))

	dst, err := os.Create(raw_file_path)
	if err != nil {
		http.Error(w, "create raw file error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "save raw file error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// http_book_get_raw_api_handler 获取书籍原始文件
// @Summary      获取书籍原始文件
// @Description  获取已有的书籍原始文件。
// @Tags         书籍
// @Accept       json
// @Produce      application/octet-stream
// @Param        book_id  path      uint64          true  "书籍ID"
// @Success      200
// @Router       /book/raw/{book_id} [get]
// @Security     Bearer
func http_book_get_raw_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err         error
		book_id_str = r.PathValue("book_id")
	)

	book_id, err := strconv.ParseUint(book_id_str, 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// 检查书籍是否存在
	err = s.DB.CheckBookID(book_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 检查书籍类型
	db_book, err := s.DB.GetBookByID(book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 查找书籍原始文件
	book_raw_path := ""
	book_path := path.Join(books_path, book_id_str)
	files, err := os.ReadDir(book_path)
	if err != nil {
		http.Error(w, "read book directory error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), "raw.") && book_raw_file_ext_check(db_book.Type, f.Name()) {
			book_raw_path = path.Join(book_path, f.Name())
			break
		}
	}

	if book_raw_path == "" {
		http.Error(w, "raw file not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, book_raw_path)
}

// http_book_pre_process_raw_api_handler 预处理书籍原始文件
// @Summary      预处理书籍原始文件
// @Description  预处理导入的书籍原始文件。
// @Tags         书籍
// @Produce      application/json
// @Param        book_id  path      uint64          true  "书籍ID"
// @Param        file     formData  file            true  "原始文件"
// @Success      200
// @Router       /book/pre_process_raw/{book_id} [post]
// @Security     Bearer
func http_book_pre_process_raw_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err         error
		book_id_str = r.PathValue("book_id")
	)

	book_id, err := strconv.ParseUint(book_id_str, 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// 检查书籍是否存在
	err = s.DB.CheckBookID(book_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 检查书籍类型
	db_book, err := s.DB.GetBookByID(book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 查找书籍原始文件
	book_raw_path := ""
	book_path := path.Join(books_path, book_id_str)
	files, err := os.ReadDir(book_path)
	if err != nil {
		http.Error(w, "read book directory error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), "raw.") && book_raw_file_ext_check(db_book.Type, f.Name()) {
			book_raw_path = path.Join(book_path, f.Name())
			break
		}
	}

	if book_raw_path == "" {
		http.Error(w, "raw file not found", http.StatusNotFound)
		return
	}

	switch db_book.Type {
	case database.WebNovelBookType:
		/***********************************************************************
		 * 对网文类型的书籍进行预处理
		 **********************************************************************/

		raw_file_buf, err := os.ReadFile(book_raw_path)
		if err != nil {
			http.Error(w, "read raw file error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		utf8_file_buf, err := web_novel_book.ConvertAnyToUTF8(raw_file_buf)
		if err != nil {
			http.Error(w, "convert raw file to UTF-8 error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 保存转换后的 UTF-8 文件到书籍目录
		err = os.WriteFile(path.Join(book_path, "utf8_raw.txt"), utf8_file_buf, 0644)
		if err != nil {
			http.Error(w, "write UTF-8 file error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		chapters, err := web_novel_book.SplitChapter(utf8_file_buf)
		if err != nil {
			http.Error(w, "split chapters error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		chapters_dir_path := path.Join(book_path, "chapters")

		if _, err := os.Stat(chapters_dir_path); err == nil {
			os.RemoveAll(chapters_dir_path)
		}

		err = s.DB.DeleteBookChaptersByBookID(book_id)
		if err != nil {
			http.Error(w, "delete book chapters error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		os.MkdirAll(chapters_dir_path, 0755)

		for i, chapter := range chapters {

			err := s.DB.CreateWebNovelChapter(&database.WebNovelChapter{
				BookID: db_book.ID,
				Index:  uint64(i),
				Name:   chapter.Title,
			})
			if err != nil {
				http.Error(w, "create chapter error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			book_chapter_path := path.Join(book_path, "chapters", strconv.FormatUint(uint64(i), 10)+".txt")

			err = os.WriteFile(book_chapter_path, []byte(chapter.Content), 0644)
			if err != nil {
				http.Error(w, "create chapter error: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

	default:
		http.Error(w, "unsupported book type for pre processing", http.StatusBadRequest)
		return
	}
}

// http_book_get_chapter_list_api_handler 获取书籍章节列表
// @Summary      获取书籍章节列表
// @Description  获取指定书籍的章节列表。
// @Tags         书籍
// @Produce      application/json
// @Param        book_id  path      uint64          true  "书籍ID"
// @Success      200      {array}   database.WebNovelChapter
// @Router       /book/chapters/{book_id} [get]
// @Security     Bearer
func http_book_get_chapter_list_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err         error
		book_id_str = r.PathValue("book_id")
	)

	book_id, err := strconv.ParseUint(book_id_str, 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// 检查书籍是否存在
	err = s.DB.CheckBookID(book_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 检查书籍类型
	db_book, err := s.DB.GetBookByID(book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch db_book.Type {
	case database.WebNovelBookType:
		/***********************************************************************
		 * 网文类型
		 **********************************************************************/
		chapters, err := s.DB.GetWebNovelChaptersByBookID(book_id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := []Chapter{}
		for _, chapter := range chapters {
			resp = append(resp, Chapter{
				Index:  chapter.Index,
				BookID: chapter.BookID,
				Title:  chapter.Name,
			})
		}

		s.WriteJsonSuccessResponse(w, resp)

	default:
		http.Error(w, "unsupported book type for getting chapter list", http.StatusBadRequest)
		return
	}

}

// http_book_get_chapter_content_api_handler 获取书籍章节内容
// @Summary      获取书籍章节内容
// @Description  获取指定书籍的章节内容。
// @Tags         书籍
// @Produce      application/json
// @Param        book_id     path      uint64          true  "书籍ID"
// @Param        index       path      uint64          true  "章节索引"
// @Success      200         {object}  database.WebNovelChapter
// @Router       /book/chapters/{book_id}/{index} [get]
// @Security     Bearer
func http_book_get_chapter_content_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		book_id_str = r.PathValue("book_id")
		index_str   = r.PathValue("index")
	)

	book_id, err := strconv.ParseUint(book_id_str, 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	index, err := strconv.ParseUint(index_str, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chapter index", http.StatusBadRequest)
		return
	}
	// 检查书籍是否存在
	err = s.DB.CheckBookID(book_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 检查书籍类型
	db_book, err := s.DB.GetBookByID(book_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	book_path := path.Join(books_path, book_id_str)

	switch db_book.Type {
	case database.WebNovelBookType:
		/***********************************************************************
		 * 网文类型
		 **********************************************************************/
		chapter, err := s.DB.GetWebNovelChaptersBookIndex(book_id, index)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if chapter == nil {
			http.Error(w, "Chapter not found", http.StatusNotFound)
			return
		}

		book_chapter_path := path.Join(book_path, "chapters", strconv.FormatUint(chapter.Index, 10)+".txt")

		http.ServeFile(w, r, book_chapter_path)

	default:
		http.Error(w, "unsupported book type for getting chapter list", http.StatusBadRequest)
		return
	}

}
