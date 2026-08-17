package main

import "api-gateway/internal/server"

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.

// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth
func main() {
	server.RunServer()
}
