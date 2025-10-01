package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	db        *sql.DB
	templates = template.Must(template.ParseGlob("templates/*.html"))
)

func main() {
	var err error
	db, err = sql.Open("sqlite3", "messenger.db")
	if err != nil {
		log.Fatal(err)
	}
	initDB()

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/chat", chatHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/test", testHandler)
	http.HandleFunc("/messages", messagesHandler)
	http.HandleFunc("/", indexHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("сайт доступен и запущен по адресу: localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func initDB() {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT,
		avatar TEXT
	);
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		content TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("DB init error:", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodPost {
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")

		log.Printf("Register attempt with username: '%s'", username)

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Ошибка при хэшировании пароля", http.StatusInternalServerError)
			return
		}
		_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, hash)
		if err == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		} else {
			templates.ExecuteTemplate(w, "register.html", map[string]string{"Error": "Имя пользователя занято или другая ошибка"})
			return
		}
	}
	templates.ExecuteTemplate(w, "register.html", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodPost {
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")

		log.Printf("Login attempt with username: '%s'", username)

		var hash string
		err := db.QueryRow("SELECT password FROM users WHERE username=?", username).Scan(&hash)
		if err == nil && bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
			cookieValue := url.QueryEscape(username)
			http.SetCookie(w, &http.Cookie{
				Name:     "user",
				Value:    cookieValue,
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			})
			http.Redirect(w, r, "/chat", http.StatusSeeOther)
			return
		} else {
			templates.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Неверное имя пользователя или пароль"})
			return
		}
	}
	templates.ExecuteTemplate(w, "login.html", nil)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user := getUserFromCookie(r)
	if user == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		message := r.FormValue("message")
		var userID int
		db.QueryRow("SELECT id FROM users WHERE username=?", user).Scan(&userID)
		if message != "" {
			db.Exec("INSERT INTO messages (user_id, content) VALUES (?, ?)", userID, message)
		}
	}

	rows, _ := db.Query(`
		SELECT u.username, u.avatar, m.content
		FROM messages m
		JOIN users u ON m.user_id = u.id
		ORDER BY m.timestamp DESC LIMIT 20`)
	defer rows.Close()

	type Message struct {
		Username string
		Avatar   string
		Content  string
	}
	var messages []Message
	for rows.Next() {
		var msg Message
		rows.Scan(&msg.Username, &msg.Avatar, &msg.Content)
		if msg.Avatar == "" {
			msg.Avatar = "/static/default-avatar.png"
		}
		messages = append(messages, msg)
	}
	templates.ExecuteTemplate(w, "chat.html", map[string]interface{}{
		"Username": user,
		"Messages": messages,
	})
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	username := getUserFromCookie(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var avatar string
	db.QueryRow("SELECT avatar FROM users WHERE username=?", username).Scan(&avatar)

	if r.Method == http.MethodPost {
		newUsername := strings.TrimSpace(r.FormValue("username"))

		file, handler, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			os.MkdirAll("static/avatars", 0755)
			path := "static/avatars/" + handler.Filename
			dst, err := os.Create(path)
			if err == nil {
				defer dst.Close()
				io.Copy(dst, file)
				db.Exec("UPDATE users SET avatar=? WHERE username=?", "/"+path, username)
				avatar = "/" + path
			}
		}

		if newUsername != "" && newUsername != username {
			_, err := db.Exec("UPDATE users SET username=? WHERE username=?", newUsername, username)
			if err == nil {
				cookieValue := url.QueryEscape(newUsername)
				http.SetCookie(w, &http.Cookie{
					Name:     "user",
					Value:    cookieValue,
					Path:     "/",
					Expires:  time.Now().Add(24 * time.Hour),
					HttpOnly: true,
				})
				username = newUsername
			} else {
				templates.ExecuteTemplate(w, "profile.html", map[string]interface{}{
					"Username": username,
					"Avatar":   avatar,
					"Error":    "Не удалось изменить имя пользователя",
				})
				return
			}
		}
	}

	templates.ExecuteTemplate(w, "profile.html", map[string]interface{}{
		"Username": username,
		"Avatar":   avatar,
	})
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user := getUserFromCookie(r)
	if user == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	templates.ExecuteTemplate(w, "test.html", map[string]interface{}{
		"Username": user,
	})
}

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	rows, err := db.Query(`
		SELECT u.username, u.avatar, m.content
		FROM messages m
		JOIN users u ON m.user_id = u.id
		ORDER BY m.timestamp DESC LIMIT 20`)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Message struct {
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
		Content  string `json:"content"`
	}
	var messages []Message
	for rows.Next() {
		var msg Message
		rows.Scan(&msg.Username, &msg.Avatar, &msg.Content)
		if msg.Avatar == "" {
			msg.Avatar = "/static/default-avatar.png"
		}
		messages = append(messages, msg)
	}
	// Отправляем JSON
	jsonData, _ := json.Marshal(messages)
	w.Write(jsonData)
}

func getUserFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("user")
	if err != nil {
		return ""
	}
	username, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return ""
	}
	return username
}
