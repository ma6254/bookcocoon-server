package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ma6254/bookcocoon-server/validator"
)

var (
	HttpErrorUploadNotFound = func(w http.ResponseWriter) {
		http.Error(w, "Upload not found", http.StatusNotFound)
	}
)

// UploadInfoRequest 上传文件信息请求
type UploadInfoRequest struct {
	Hash string `json:"hash"` // 文件哈希
	Name string `json:"name"` // 文件名
	Size uint64 `json:"size"` // 文件大小
}

type UploadInfoResponse struct {
	FileID     string `json:"file_id"`     // 文件ID
	Hash       string `json:"hash"`        // 文件哈希SHA256
	Name       string `json:"name"`        // 文件名
	Size       uint64 `json:"size"`        // 文件大小
	Path       string `json:"path"`        // 文件存储路径
	CreatedAt  string `json:"created_at"`  // 创建日期
	UploadedAt string `json:"uploaded_at"` // 上传日期
	UploaderID uint64 `json:"uploader_id"` // 上传者ID
}

// http_create_upload_api_handler 创建上传文件信息
// @Summary      创建上传文件信息
// @Description  创建上传文件信息。
// @Tags         上传文件
// @Accept       json
// @Produce      json
// @Param        body  body      UploadInfoRequest  true  "上传文件信息请求"
// @Success      200  {object}  UploadInfoResponse
// @Router       /upload/create [post]
// @Security     Bearer
func http_create_upload_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err error
		req = UploadInfoRequest{}
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		HttpErrorJsonDecode(w, err)
		return
	}

	// 检查hash是否符合sha256
	if err := validator.ValidateSha256Hex(req.Hash); err != nil {
		http.Error(w, "hash Validate fail", http.StatusBadRequest)
		return
	}

	// 统一使用小写，避免大小写差异导致哈希比较失败
	req.Hash = strings.ToLower(req.Hash)

	// 检查文件名是否有效
	if err := validator.ValidateUploadFileName(req.Name); err != nil {
		http.Error(w, "file name validate fail", http.StatusBadRequest)
		return
	}

	// 检查是否已存在相同哈希的文件
	existingUpload, err := s.DB.FindUploadByHash(req.Hash)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}

	// 文件已存在，返回现有文件信息
	if existingUpload != nil {

		s.WriteJsonSuccessResponse(w, UploadInfoResponse{
			FileID:     fmt.Sprintf("%d", existingUpload.FileID),
			Hash:       existingUpload.Hash,
			Name:       existingUpload.Name,
			Size:       existingUpload.Size,
			Path:       existingUpload.Path,
			CreatedAt:  existingUpload.CreatedAt,
			UploadedAt: existingUpload.UploadedAt,
			UploaderID: existingUpload.CreatedByID,
		})
		return
	}

	uploaded_time := time.Now()

	// 文件保存路径
	upload_path := path.Join(
		fmt.Sprintf("%04d", uploaded_time.Year()),
		fmt.Sprintf("%02d", uploaded_time.Month()),
		fmt.Sprintf("%02d", uploaded_time.Day()),
		fmt.Sprintf("%X_%s", req.Hash[:8], req.Name),
	)

	file_id := int64(s.snowflake_node.Generate())

	session.log.Printf("Creating upload: id=%d, hash=%#v, name=%#v, size=%d, path=%#v, uploader_id=%d", file_id, req.Hash, req.Name, req.Size, upload_path, session.UserID)

	upload, err := s.DB.CreateUpload(file_id, req.Hash, req.Name, req.Size, upload_path, session.UserID)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}
	if upload == nil {
		http.Error(w, "Upload already exists", http.StatusConflict)
		return
	}

	s.WriteJsonSuccessResponse(w, UploadInfoResponse{
		FileID:     fmt.Sprintf("%d", file_id), // 文件ID
		Hash:       upload.Hash,
		Name:       upload.Name,
		Size:       upload.Size,
		Path:       upload.Path,
		CreatedAt:  upload.CreatedAt,
		UploadedAt: upload.UploadedAt,
		UploaderID: upload.CreatedByID,
	})
}

// http_upload_data_api_handler 上传文件数据
// @Summary      上传文件数据
// @Description  上传文件数据。
// @Tags         上传文件
// @Accept       multipart/form-data
// @Produce      json
// @Param        file_id  path      string  true  "文件ID"
// @Param        file     formData  file    true  "上传的文件"
// @Success      200  {object}  UploadInfoResponse
// @Router       /upload/data/{file_id} [post]
// @Security     Bearer
func http_upload_data_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		err         error
		file_id_str = r.PathValue("file_id")
	)

	file_id, err := strconv.ParseInt(file_id_str, 10, 64)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	// 检查是否已存在相同哈希的文件
	upload_info, err := s.DB.FindUploadByID(file_id)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}
	if upload_info == nil {
		HttpErrorUploadNotFound(w)
		return
	}

	var (
		fileReader = r.Body
	)

	upload_info.Size = uint64(r.ContentLength)

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		file, _, fileErr := r.FormFile("file")
		if fileErr != nil {
			WriteHttpErrorWithCode(w, session, http.StatusBadRequest, "Failed to read file from form: %v", fileErr)
			return
		}
		defer file.Close()
		fileReader = file
	}

	tmp_path := path.Join("uploads", "tmp")
	err = os.MkdirAll(tmp_path, os.ModePerm)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}

	err = s.saveUploadFile(upload_info, fileReader)
	if err != nil {
		session.log.Printf("Error saving upload file: %v", err)
		HttpErrorInternal(w, err)
		return
	}

	_, err = s.DB.UpdateUploadSize(file_id, upload_info.Size)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}

	_, err = s.DB.UpdateUploadTime(file_id)
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}

	s.WriteJsonSuccessResponse(w, UploadInfoResponse{
		FileID:     fmt.Sprintf("%d", upload_info.FileID),
		Hash:       upload_info.Hash,
		Name:       upload_info.Name,
		Size:       upload_info.Size,
		Path:       upload_info.Path,
		CreatedAt:  upload_info.CreatedAt,
		UploadedAt: upload_info.UploadedAt,
		UploaderID: upload_info.CreatedByID,
	})
}

// http_upload_read_api_handler 读取上传文件
// @Summary      读取上传文件
// @Description  读取上传文件。
// @Tags         文件
// @Accept       json
// @Produce      json
// @Param        file_id   path      string  true  "上传文件ID" allowReserved(true)
// @Success      200  {file}  octet-stream
// @Router       /upload/data/{file_id} [get]
// @Security     Bearer
func http_upload_read_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var (
		file_path = r.PathValue("file_id")
	)

	session.log.Printf("Reading upload file: path=%#v", file_path)

	file_id, err := strconv.ParseInt(file_path, 10, 64)
	if err == nil {

		// session.log.Printf("is numeric file_id: %d", file_id)

		// 如果路径是数字，尝试按文件ID查找
		upload_info, err := s.DB.FindUploadByID(file_id)
		if err != nil {
			HttpErrorInternal(w, err)
			return
		}

		// 文件不存在
		if upload_info == nil {
			HttpErrorUploadNotFound(w)
			return
		}

		src_file_path := path.Join("uploads", upload_info.Path)

		// 文件存在，返回文件内容
		file, err := os.Open(src_file_path)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()

		io.Copy(w, file)
		return
	}

	src_file_path := path.Join("uploads", file_path)

	// 文件存在，返回文件内容
	file, err := os.Open(src_file_path)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	io.Copy(w, file)
}

// http_upload_list_api_handler 上传文件列表
// @Summary      上传文件列表
// @Description  上传文件列表。
// @Tags         上传, 文件
// @Accept       json
// @Produce      json
// @Success      200  {object} []UploadInfoResponse
// @Router       /upload/list [get]
// @Security     Bearer
func http_upload_list_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	uploads, err := s.DB.GetAllUploads()
	if err != nil {
		HttpErrorInternal(w, err)
		return
	}

	var response []UploadInfoResponse
	for _, upload := range uploads {
		response = append(response, UploadInfoResponse{
			FileID:     fmt.Sprintf("%d", upload.FileID),
			Hash:       upload.Hash,
			Name:       upload.Name,
			Size:       upload.Size,
			Path:       upload.Path,
			CreatedAt:  upload.CreatedAt,
			UploadedAt: upload.UploadedAt,
			UploaderID: upload.CreatedByID,
		})
	}

	s.WriteJsonSuccessResponse(w, response)
}
