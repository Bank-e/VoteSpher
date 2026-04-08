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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func ConnectDB() *gorm.DB {
	host     := os.Getenv("DB_HOST")
	port     := os.Getenv("DB_PORT")
	user     := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname   := os.Getenv("DB_NAME")
	caCert   := os.Getenv("DB_CA_CERT")

	// โหลด CA Certificate
	rootCertPool := x509.NewCertPool()
	pem, err := os.ReadFile(caCert)
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
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
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	log.Println("Database connected successfully")
	return db
}