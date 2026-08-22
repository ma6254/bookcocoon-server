package server

import (
	_ "github.com/ma6254/bookcocoon-server/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	api_url_prefix = "/api"
)

func (s *Server) setRoute() error {

	// Swagger
	s.Mux.HandleFunc("/swagger/", httpSwagger.Handler())

	// User
	s.Mux.HandleFunc("POST "+api_url_prefix+"/user/login", s.http_api_login_handler())

	// Upload
	s.HandleTokenFunc("POST "+api_url_prefix+"/upload/create", http_create_upload_api_handler)
	s.HandleTokenFunc("POST "+api_url_prefix+"/upload/data/{file_id}", http_upload_data_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/upload/data/{file_id}", http_upload_read_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/upload/list", http_upload_list_api_handler)

	// Sys
	s.HandleTokenFunc("GET "+api_url_prefix+"/sys/info", http_sys_info_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/sys/state", http_sys_state_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/os/info", http_os_info_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/os/state", http_os_state_api_handler)

	// Book
	s.HandleTokenFunc("POST "+api_url_prefix+"/book/create", http_book_create_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/book/list", http_book_list_api_handler)
	s.HandleTokenFunc("POST "+api_url_prefix+"/book/update/{book_id}", http_book_update_api_handler)
	s.HandleTokenFunc("POST "+api_url_prefix+"/book/cover/{book_id}", http_book_update_cover_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/book/cover/{book_id}", http_book_get_cover_api_handler)
	s.HandleTokenFunc("POST "+api_url_prefix+"/book/raw/{book_id}", http_book_update_raw_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/book/raw/{book_id}", http_book_get_raw_api_handler)
	s.HandleTokenFunc("POST "+api_url_prefix+"/book/pre_process_raw/{book_id}", http_book_pre_process_raw_api_handler)

	// Book Chapters
	s.HandleTokenFunc("GET "+api_url_prefix+"/book/chapters/{book_id}", http_book_get_chapter_list_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/book/chapters/{book_id}/{index}", http_book_get_chapter_content_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/book/chapter-info/{book_id}/{index}", http_book_get_chapter_info_api_handler)

	// Reading Record
	s.HandleTokenFunc("POST "+api_url_prefix+"/reading_record/create", http_reading_record_create_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/reading_record/{book_id}", http_reading_record_get_api_handler)
	s.HandleTokenFunc("POST "+api_url_prefix+"/reading_record/{book_id}", http_reading_record_update_api_handler)
	s.HandleTokenFunc("GET "+api_url_prefix+"/reading_record", http_reading_record_list_by_user_api_handler)

	return nil
}
