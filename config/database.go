package config

import (
    "crypto/tls"
    "crypto/x509"
    "database/sql"
    "fmt"
    "log"
    "os"

    "github.com/go-sql-driver/mysql"
    "github.com/joho/godotenv"
)

func LoadEnv() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
}

func ConnectDB() *sql.DB {
    host     := os.Getenv("DB_HOST")
    port     := os.Getenv("DB_PORT")
    user     := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname   := os.Getenv("DB_NAME")
    caCert   := os.Getenv("DB_CA_CERT")

    // อ่าน CA Certificate จากไฟล์
    rootCertPool := x509.NewCertPool()
    pem, err := os.ReadFile(caCert)
    if err != nil {
        log.Fatalf("Failed to read CA cert: %v", err)
    }
    if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
        log.Fatal("Failed to append CA cert")
    }

    // Register TLS config
    mysql.RegisterTLSConfig("aiven", &tls.Config{
        RootCAs: rootCertPool,
    })

    // DSN ใช้ tls=aiven แทน tls=true
    dsn := fmt.Sprintf(
        "%s:%s@tcp(%s:%s)/%s?tls=aiven&parseTime=true",
        user, password, host, port, dbname,
    )

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Failed to open DB: %v", err)
    }

    if err := db.Ping(); err != nil {
        log.Fatalf("Failed to connect to DB: %v", err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)

    log.Println("Database connected successfully")
    return db
}