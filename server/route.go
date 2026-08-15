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
	s.Mux.HandleFunc(api_url_prefix+"/user/login", s.http_api_login_handler())

	return nil
}
