package main

import "github.com/shopicano/shopicano-backend/cmd"

// @title Shopicano Backend API
// @version 1.0
// @description Shopicano backend RESTful API service
// @termsOfService https://www.shopicano.com/terms

// @contact.name Shopicano Developers
// @contact.url https://www.shopicano.com
// @contact.email developers@shopicano.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:9119
// @BasePath /v1

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	cmd.Execute()
}
