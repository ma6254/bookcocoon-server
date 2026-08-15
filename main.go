/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
// @title           BookCocoon Server API
// @version        1.0
// @description    BookCocoon Server API，使用 Swagger 文档生成工具。
// @termsOfService http://swagger.io/terms/

// @contact.name   API 支持
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      127.0.0.1:28080
// @BasePath  /api

// @securityDefinitions.apikey  Bearer
// @in header
// @name Authorization
package main

import "github.com/ma6254/bookcocoon-server/cmd"

func main() {
	cmd.Execute()
}
