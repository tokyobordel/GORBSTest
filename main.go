package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

type Message struct {
	ID        int          `json:"id"`
	Content   string       `json:"content"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
	DeletedAt sql.NullTime `json:"-"` // не сериализуем напрямую
}

type messageJSON struct {
	ID        int        `json:"id"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// Преобразование Message -> messageJSON
func toJSON(m Message) messageJSON {
	mj := messageJSON{
		ID:        m.ID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
	if m.UpdatedAt.Valid {
		mj.UpdatedAt = &m.UpdatedAt.Time
	}
	if m.DeletedAt.Valid {
		mj.DeletedAt = &m.DeletedAt.Time
	}
	return mj
}

var upgrader = websocket.Upgrader{
	// Разрешаем соединения с любых источников (в продакшене лучше ограничить)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var db *sql.DB

// Список клиентов. Клиенты нужны для рассылки вновь добавленных сообщений
var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "123"),
		getEnv("DB_NAME", "postgres"),
		getEnv("DB_SSLMODE", "disable"),
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Не удалось открыть БД: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Нет соединения с БД: %v", err)
	}

	createTable := "CREATE TABLE IF NOT EXISTS messages " +
		"( id SERIAL PRIMARY KEY, " +
		"content TEXT NOT NULL, " +
		"created_at TIMESTAMP NOT NULL DEFAULT LOCALTIMESTAMP, " +
		"updated_at TIMESTAMP, " +
		"deleted_at TIMESTAMP );"

	db.Exec(createTable)

	r := mux.NewRouter()
	r.HandleFunc("/chat", handleChatWS)
	r.HandleFunc("/chat_history", handleHistoryWS)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	port := ":8080"
	log.Fatal(http.ListenAndServe(port, r))
}

// handleChatWS принимает сообщения и сохраняет их в БД.
func handleChatWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Фиксируем соединение (тобишь клиента)
	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
	}()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Ожидаем JSON с полем "content"
		var incoming struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(msgBytes, &incoming); err != nil {
			log.Printf("Невалидный JSON в /chat: %v", err)
			continue
		}
		if incoming.Content == "" {
			continue
		}

		// Вставка в БД с возвратом всех полей
		var msg Message
		err = db.QueryRow(
			`INSERT INTO messages (content, created_at)
			 VALUES ($1, NOW())
			 RETURNING id, content, created_at, updated_at, deleted_at`,
			incoming.Content,
		).Scan(&msg.ID, &msg.Content, &msg.CreatedAt, &msg.UpdatedAt, &msg.DeletedAt)
		if err != nil {
			continue
		}

		broadcast()
	}
}

func loadAllMessages() ([]messageJSON, error) {
	rows, err := db.Query(
		`SELECT id, content, created_at, updated_at, deleted_at
         FROM messages
         WHERE deleted_at IS NULL
         ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []messageJSON
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Content, &msg.CreatedAt, &msg.UpdatedAt, &msg.DeletedAt); err != nil {
			continue
		}
		history = append(history, toJSON(msg))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return history, nil
}

func broadcast() {
	messages, err := loadAllMessages()
	if err != nil {
		return
	}
	data, _ := json.Marshal(messages)

	clientsMu.Lock()
	defer clientsMu.Unlock()
	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			delete(clients, conn)
		}
	}
}

// handleHistoryWS отправляет всю историю активных сообщений.
func handleHistoryWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		// Ждём любое сообщение от клиента (например, "ping")
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}

		messages, _ := loadAllMessages()

		data, _ := json.Marshal(messages)

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			break
		}
	}
}
