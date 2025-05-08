package controllers

import (
	"AutoM/config"
	"AutoM/models"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// -----------------------------
// JSON и шаблонные утилиты
// -----------------------------

// JSONResponse – универсальная структура для отправки JSON-ответов.
type JSONResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// sendJSON централизует отправку JSON-ответа.
func sendJSON(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Ошибка отправки JSON-ответа: %v", err)
	}
}

// tplCache хранит распарсенные шаблоны для повторного использования.
var tplCache = make(map[string]*template.Template)

// var tplOnce sync.Once

// RenderTemplateCached выполняет кэширование и рендеринг HTML-шаблона.
func RenderTemplateCached(w http.ResponseWriter, tmplName string, data interface{}) {
	tplPath := filepath.Join("templates", tmplName)
	tpl, ok := tplCache[tmplName]
	if !ok {
		var err error
		tpl, err = template.ParseFiles(tplPath)
		if err != nil {
			log.Printf("[ERROR] Не удалось загрузить шаблон %s: %v", tplPath, err)
			http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
			return
		}
		tplCache[tmplName] = tpl
	}
	if err := tpl.Execute(w, data); err != nil {
		log.Printf("[ERROR] Ошибка исполнения шаблона %s: %v", tplPath, err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// -----------------------------
// Сессионное управление
// -----------------------------

// getSession пытается получить сессию. Если обнаруживается ошибка вида
// "securecookie: the value is not valid", создаётся новая сессия.
func getSession(_ http.ResponseWriter, r *http.Request) (*sessions.Session, error) {
	session, err := config.Store.Get(r, "session")
	if err != nil && err.Error() == "securecookie: the value is not valid" {
		log.Printf("[WARN] Обнаружен недействительный cookie сессии, создается новая.")
		session, _ = config.Store.New(r, "session")
		return session, nil
	}
	return session, err
}

// -----------------------------
// Обработчики пользователей
// -----------------------------

// RegisterUserHandler регистрирует нового пользователя.
func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Printf("[ERROR] Ошибка декодирования JSON для регистрации: %v", err)
		sendJSON(w, http.StatusBadRequest, JSONResponse{Error: "Неверный формат данных"})
		return
	}

	reqData.Username = strings.TrimSpace(reqData.Username)
	reqData.Email = strings.TrimSpace(reqData.Email)
	reqData.Password = strings.TrimSpace(reqData.Password)
	if reqData.Username == "" || reqData.Email == "" || reqData.Password == "" {
		sendJSON(w, http.StatusBadRequest, JSONResponse{Error: "Все поля должны быть заполнены"})
		return
	}
	if len(reqData.Password) < 6 {
		sendJSON(w, http.StatusBadRequest, JSONResponse{Error: "Пароль должен содержать не менее 6 символов"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqData.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] Ошибка хэширования пароля: %v", err)
		sendJSON(w, http.StatusInternalServerError, JSONResponse{Error: "Ошибка обработки пароля"})
		return
	}

	err = RegisterUser(config.DB, reqData.Username, reqData.Email, string(hashedPassword))
	if err != nil {
		log.Printf("[ERROR] Ошибка регистрации пользователя (%s): %v", reqData.Username, err)
		sendJSON(w, http.StatusInternalServerError, JSONResponse{Error: "Ошибка регистрации пользователя"})
		return
	}

	log.Printf("[INFO] Пользователь %s успешно зарегистрирован", reqData.Username)
	sendJSON(w, http.StatusCreated, JSONResponse{Message: "Пользователь успешно зарегистрирован"})
}

// AuthorizeUserHandler авторизует пользователя, устанавливая в сессии значения:
// "user_id", "username" и "is_admin". Здесь используется config.Store.
func AuthorizeUserHandler(w http.ResponseWriter, r *http.Request) {
	var reqData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Printf("[ERROR] Ошибка декодирования JSON для авторизации: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	reqData.Username = strings.TrimSpace(reqData.Username)
	reqData.Password = strings.TrimSpace(reqData.Password)
	if reqData.Username == "" || reqData.Password == "" {
		http.Error(w, "Поля логина и пароля обязательны", http.StatusBadRequest)
		return
	}

	user, err := GetUserByUsername(config.DB, reqData.Username)
	if err != nil {
		log.Printf("[ERROR] Пользователь не найден для имени %s: %v", reqData.Username, err)
		http.Error(w, "Неверные учетные данные", http.StatusUnauthorized)
		return
	}
	log.Printf("[DEBUG] Пользователь найден: Username=%s, IsAdmin=%v", user.Username, user.IsAdmin)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqData.Password)); err != nil {
		log.Printf("[ERROR] Неверный пароль для пользователя %s: %v", reqData.Username, err)
		http.Error(w, "Неверные учетные данные", http.StatusUnauthorized)
		return
	}

	// Используем config.Store вместо неопределенной переменной store.
	session, err := config.Store.Get(r, "session")
	if err != nil {
		log.Printf("[ERROR] Не удалось получить сессию: %v", err)
		http.Error(w, "Ошибка обработки сессии", http.StatusInternalServerError)
		return
	}
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Values["is_admin"] = user.IsAdmin
	log.Printf("[DEBUG] Установка сессии: UserID=%v, Username=%s, IsAdmin=%v",
		session.Values["user_id"], session.Values["username"], session.Values["is_admin"])
	if err = session.Save(r, w); err != nil {
		log.Printf("[ERROR] Не удалось сохранить сессию: %v", err)
		http.Error(w, "Ошибка сохранения сессии", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] Пользователь %s успешно авторизовался", reqData.Username)
	acceptHeader := r.Header.Get("Accept")
	if strings.Contains(acceptHeader, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		resp := JSONResponse{
			Message: "Авторизация успешна",
			Data: map[string]interface{}{
				"username": user.Username,
				"is_admin": user.IsAdmin,
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("[ERROR] Ошибка кодирования JSON-ответа: %v", err)
		}
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// PersonalCabinetHandler отображает страницу личного кабинета.
// Если пользователь не авторизован, перенаправляет на страницу авторизации.
func PersonalCabinetHandler(w http.ResponseWriter, r *http.Request) {
	session, err := config.Store.Get(r, "session")
	if err != nil {
		log.Printf("[ERROR] Не удалось получить сессию: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	username, ok := session.Values["username"].(string)
	if !ok || username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := GetUserByUsername(config.DB, username)
	if err != nil {
		log.Printf("[ERROR] Ошибка получения данных пользователя %s: %v", username, err)
		http.Error(w, "Ошибка получения данных", http.StatusInternalServerError)
		return
	}
	tplPath := filepath.Join("templates", "cabinet.html")
	tpl, err := template.ParseFiles(tplPath)
	if err != nil {
		log.Printf("[ERROR] Ошибка загрузки шаблона личного кабинета %s: %v", tplPath, err)
		http.Error(w, "Ошибка загрузки страницы", http.StatusInternalServerError)
		return
	}
	if err := tpl.Execute(w, user); err != nil {
		log.Printf("[ERROR] Ошибка исполнения шаблона личного кабинета: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}

// LogoutHandler очищает сессию и перенаправляет пользователя на главную страницу.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := config.Store.Get(r, "session")
	if err != nil {
		log.Printf("[ERROR] Не удалось получить сессию: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	session.Values["user_id"] = nil
	session.Values["username"] = nil
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		log.Printf("[ERROR] Ошибка очистки сессии: %v", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// RegisterUser сохраняет нового пользователя в базу.
// При регистрации для всех пользователей is_admin выставляется в false.
func RegisterUser(db *sql.DB, username, email, hashedPassword string) error {
	query := `
        INSERT INTO users (username, email, password, is_admin, created_at, updated_at)
        VALUES (?, ?, ?, ?, NOW(), NOW())
    `
	// Передаём false для is_admin (в MySQL это обычно 0).
	_, err := db.Exec(query, username, email, hashedPassword, false)
	return err
}

// GetUserByUsername выполняет выборку пользователя по username.
func GetUserByUsername(db *sql.DB, username string) (models.User, error) {
	var user models.User
	query := "SELECT id, username, email, password, is_admin FROM users WHERE username = ?"
	err := db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.IsAdmin,
	)
	if err != nil {
		return models.User{}, err
	}

	// Логирование для проверки значения is_admin
	log.Printf("[DEBUG] Данные из базы: Username=%s, IsAdmin=%v", user.Username, user.IsAdmin)
	return user, nil
}

// GetAllUsers – возвращает список всех пользователей в формате JSON.
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] GetAllUsers: Начало запроса пользователей")
	query := "SELECT id, username, email, is_admin FROM users"
	log.Printf("[INFO] GetAllUsers: Выполняется запрос: %s", query)
	rows, err := config.DB.Query(query)
	if err != nil {
		log.Printf("[ERROR] GetAllUsers: Ошибка выполнения запроса: %v", err)
		http.Error(w, "Ошибка загрузки пользователей", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("[ERROR] GetAllUsers: Ошибка закрытия rows: %v", err)
		} else {
			log.Println("[INFO] GetAllUsers: rows успешно закрыты")
		}
	}()

	var users []models.User
	rowNum := 0
	for rows.Next() {
		rowNum++
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin); err != nil {
			log.Printf("[ERROR] GetAllUsers: Ошибка сканирования строки %d: %v", rowNum, err)
			http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] GetAllUsers: Ошибка итерации по строкам: %v", err)
		http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] GetAllUsers: успешно получено %d пользователей", len(users))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("[ERROR] GetAllUsers: Ошибка кодирования JSON: %v", err)
	}
}

// GetUserByID – возвращает пользователя по его ID в формате JSON.
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] GetUserByID: Начало запроса пользователя")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("[ERROR] GetUserByID: Неверный формат ID: %v", err)
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}
	log.Printf("[INFO] GetUserByID: Получен ID: %d", id)
	query := "SELECT id, username, email, is_admin FROM users WHERE id = ?"
	log.Printf("[INFO] GetUserByID: Выполняется запрос: %s", query)
	row := config.DB.QueryRow(query, id)
	var user models.User
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin); err != nil {
		log.Printf("[ERROR] GetUserByID: Пользователь с ID %d не найден: %v", id, err)
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}
	log.Printf("[INFO] GetUserByID: Пользователь получен: %+v", user)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("[ERROR] GetUserByID: Ошибка кодирования JSON: %v", err)
	}
}
