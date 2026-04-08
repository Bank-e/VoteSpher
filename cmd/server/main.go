package main

import (
    "VoteSphere/config"
)

func main() {
    // โหลด .env ก่อนทุกอย่าง
    config.LoadEnv()

    // เชื่อมต่อ DB
    db := config.ConnectDB()
    defer db.Close()

    // ... register routes, inject db ต่อไป
}