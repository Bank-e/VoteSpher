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

// LoadEnv โหลดตัวแปรจากไฟล์ .env
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println(" .env not found, using system env instead")
	}
}

// ConnectDB เป็นฟังก์ชันหลัก (Entry point) ที่ชี้ว่าจะไป Cloud หรือ Local
func ConnectDB() *gorm.DB {
	dbMode := os.Getenv("DB_MODE")
	
	if dbMode == "cloud" {
		return ConnectCloudDB()
	}
	return ConnectLocalDB() // ค่า Default หากไม่ได้ระบุให้เป็น Local
}

// ConnectLocalDB สำหรับเชื่อมต่อ Docker/Local (ดึงค่า LOCAL_*)
func ConnectLocalDB() *gorm.DB {
	host := os.Getenv("LOCAL_DB_HOST")
	port := os.Getenv("LOCAL_DB_PORT")
	user := os.Getenv("LOCAL_DB_USER")
	password := os.Getenv("LOCAL_DB_PASSWORD")
	dbname := os.Getenv("LOCAL_DB_NAME")

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		user, password, host, port, dbname,
	)

	log.Println("Connecting to LOCAL Database (Docker)...")
	return openGormConnection(dsn)
}

// ConnectCloudDB สำหรับเชื่อมต่อ Cloud/Aiven (ดึงค่า CLOUD_*)
func ConnectCloudDB() *gorm.DB {
	host := os.Getenv("CLOUD_DB_HOST")
	port := os.Getenv("CLOUD_DB_PORT")
	user := os.Getenv("CLOUD_DB_USER")
	password := os.Getenv("CLOUD_DB_PASSWORD")
	dbname := os.Getenv("CLOUD_DB_NAME")
	caCert := os.Getenv("CLOUD_DB_CA_CERT")

	if caCert == "" {
		log.Fatal("CLOUD_DB_CA_CERT is required for cloud mode")
	}

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
		"%s:%s@tcp(%s:%s)/%s?tls=aiven&parseTime=true&loc=Local",
		user, password, host, port, dbname,
	)

	log.Println("Connecting to CLOUD Database (TLS Enabled)...")
	return openGormConnection(dsn)
}

// openGormConnection เป็น Helper function จัดการเรื่องการเปิด DB และ Connection Pool
func openGormConnection(dsn string) *gorm.DB {
	db, err := gorm.Open(gormMySQL.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // แสดง SQL query ใน console
	})
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Connection Pool Settings
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	log.Println("✅ Database connected successfully")
	return db
}