package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	gormMySQL "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env not found, using system env instead")
	}
}

func ConnectDB() *gorm.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// รองรับทั้ง path ไฟล์ (local) และ content ตรงๆ (cloud/Railway)
	var pemBytes []byte
	if content := os.Getenv("DB_CA_CERT_CONTENT"); content != "" {
		pemBytes = []byte(content)
	} else {
		var err error
		pemBytes, err = os.ReadFile(os.Getenv("DB_CA_CERT"))
		if err != nil {
			log.Fatalf("Failed to read CA cert: %v", err)
		}
	}

	rootCertPool := x509.NewCertPool()
	if ok := rootCertPool.AppendCertsFromPEM(pemBytes); !ok {
		log.Fatal("Failed to append CA cert")
	}

	mysql.RegisterTLSConfig("aiven", &tls.Config{
		RootCAs: rootCertPool,
	})

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?tls=aiven&parseTime=true",
		user, password, host, port, dbname,
	)

	db, err := gorm.Open(gormMySQL.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // แสดง SQL query ใน console
	})
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Connection Pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	log.Println("Database connected successfully")
	return db
}
