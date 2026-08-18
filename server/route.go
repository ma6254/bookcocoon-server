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

	return nil
}
