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
		log.Println(" .env not found, using system env instead")
	}
}

func ConnectDB() *gorm.DB {
	dbMode := os.Getenv("DB_MODE")
	if dbMode == "cloud" {
		return ConnectCloudDB()
	}
	return ConnectLocalDB()
}

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

func ConnectCloudDB() *gorm.DB {
	host := os.Getenv("CLOUD_DB_HOST")
	port := os.Getenv("CLOUD_DB_PORT")
	user := os.Getenv("CLOUD_DB_USER")
	password := os.Getenv("CLOUD_DB_PASSWORD")
	dbname := os.Getenv("CLOUD_DB_NAME")
	caCertPath := os.Getenv("CLOUD_DB_CA_CERT")

	if caCertPath == "" {
		log.Fatal("CLOUD_DB_CA_CERT is required for cloud mode")
	}

	var pemBytes []byte
	if content := os.Getenv("DB_CA_CERT_CONTENT"); content != "" {
		pemBytes = []byte(content)
	} else {
		var err error
		pemBytes, err = os.ReadFile(caCertPath)
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
		"%s:%s@tcp(%s:%s)/%s?tls=aiven&parseTime=true&loc=Local",
		user, password, host, port, dbname,
	)

	log.Println("Connecting to CLOUD Database (TLS Enabled)...")
	return openGormConnection(dsn)
}

func openGormConnection(dsn string) *gorm.DB {
	db, err := gorm.Open(gormMySQL.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	log.Println("✅ Database connected successfully")
	return db
}
