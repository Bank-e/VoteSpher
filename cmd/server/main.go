package main

import (
	"votespher/config"
	"votespher/migration"
)

func main() {
	config.LoadEnv()

	db := config.ConnectDB()

	// รัน migration ทุกครั้งที่ start server
	migration.Run(db)

	// ... register routes ต่อไป
}