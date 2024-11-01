package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite" // 使用 pure Go SQLite 驅動避免 cgo
)

var db *sql.DB
var jwtKey = []byte("your_secret_key")
var upgrader = websocket.Upgrader{}

type User struct {
	ID       int
	Username string
	Password string
}

func main() {
	var err error
	db, err = sql.Open("sqlite", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/update-password", updatePasswordHandler) // 密碼更新路由
	http.HandleFunc("/ws", wsHandler)                          // WebSocket 路由

	log.Println("Server started on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func createTable() {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE,
        password TEXT
    );`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var message string
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		_, err := findUserByUsername(username)
		if err == nil {
			message = "Username already exists. Please choose another."
		} else {
			query := `INSERT INTO users (username, password) VALUES (?, ?)`
			_, err = db.Exec(query, username, password)
			if err != nil {
				message = "Failed to register user. Please try again."
			} else {
				message = "Registration successful! You can now log in."
			}
		}
	}

	tmpl, _ := template.ParseFiles("templates/register.html")
	tmpl.Execute(w, map[string]string{"Message": message})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var message string
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := findUserByUsername(username)
		if err != nil || user.Password != password {
			message = "Invalid credentials. Please try again."
		} else {
			token, err := generateToken(username)
			if err != nil {
				http.Error(w, "Failed to generate token", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    token,
				Expires:  time.Now().Add(1 * time.Hour),
				HttpOnly: true,
			})

			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
	}

	tmpl, _ := template.ParseFiles("templates/login.html")
	tmpl.Execute(w, map[string]string{"Message": message})
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tokenStr := cookie.Value
	claims := &jwt.StandardClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 這裡檢查用戶是否存在，並正確獲取用戶資料
	user, err := findUserByUsername(claims.Subject)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 這裡確保正確傳遞用戶 ID 和密碼
	tmpl, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		http.Error(w, "Failed to load dashboard template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Username": claims.Subject,
		"UserID":   user.ID,
		"Password": user.Password, // 添加密碼以顯示在頁面
		"Token":    tokenStr,
	})
}

func updatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tokenStr := cookie.Value
	claims := &jwt.StandardClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		newPassword := r.FormValue("newPassword")
		if newPassword == "" {
			http.Redirect(w, r, "/dashboard?error=New password cannot be empty.", http.StatusSeeOther)
			return
		}

		_, err := db.Exec(`UPDATE users SET password = ? WHERE username = ?`, newPassword, claims.Subject)
		if err != nil {
			http.Redirect(w, r, "/dashboard?error=Failed to update password.", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to set websocket upgrade:", err)
		return
	}
	defer conn.Close()

	for {
		timeNow := time.Now().Format("2006-01-02 15:04:05")
		if err := conn.WriteMessage(websocket.TextMessage, []byte(timeNow)); err != nil {
			log.Println("Error writing to websocket:", err)
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now().Add(-time.Hour),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func findUserByUsername(username string) (*User, error) {
	query := `SELECT id, username, password FROM users WHERE username = ?`
	row := db.QueryRow(query, username)

	user := &User{}
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func generateToken(username string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &jwt.StandardClaims{
		Subject:   username,
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
