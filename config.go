package config

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

const (
	// defaultDSN используется, если переменная окружения DSN не задана.
	defaultDSN = "root:Admin12345@tcp(127.0.0.1:3306)/mydb?parseTime=true"
	// defaultSecretKey – статичный секретный ключ для сессий (используйте надёжное значение в продакшене).
	defaultSecretKey = "123"
)

// DB – глобальное соединение с базой данных.
var DB *sql.DB

// Store – глобальное хранилище сессий.
var Store *sessions.CookieStore

// InitDB инициализирует подключение к базе данных.
// DSN можно задать через переменную окружения "DSN", если не установлен – используется значение по умолчанию.
// В продакшене рекомендуется использовать надёжные параметры подключения.
func InitDB() {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = defaultDSN
		log.Println("Переменная DSN не установлена; используется значение по умолчанию (НЕ для продакшена!)")
	}

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatalf("Ошибка проверки соединения с базой: %v", err)
	}
	log.Println("Успешное подключение к базе данных!")
}

// InitStore инициализирует хранилище сессий со статичным секретным ключом.
// В продакшене следует обеспечивать безопасное хранение секрета.
func InitStore() {
	Store = sessions.NewCookieStore([]byte(defaultSecretKey))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,  // Cookie действительна 1 час
		HttpOnly: true,  // Защищает cookie от доступа через JavaScript
		Secure:   false, // Установите true, если приложение работает по HTTPS
	}
	log.Println("Хранилище сессий инициализировано со статичным секретным ключом")
}

// CloseDB закрывает соединение с базой данных.
func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Ошибка при закрытии подключения к базе данных: %v", err)
		} else {
			log.Println("Соединение с базой данных закрыто.")
		}
	}
}
