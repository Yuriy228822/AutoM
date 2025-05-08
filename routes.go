package routes

import (
	"AutoM/controllers"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// LoggingFileServer оборачивает стандартный файловый сервер, логируя каждый запрос.
func LoggingFileServer(dir http.Dir) http.Handler {
	fileServer := http.FileServer(dir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] Запрос статического файла: %s", r.URL.Path)
		fileServer.ServeHTTP(w, r)
	})
}

// RegisterRoutes настраивает все маршруты приложения и возвращает корневой роутер.
func RegisterRoutes() *mux.Router {
	// Создаем основной роутер.
	router := mux.NewRouter()

	// Обслуживание статических файлов из папки "dist" с логированием.
	router.PathPrefix("/dist/").Handler(http.StripPrefix("/dist/", LoggingFileServer(http.Dir("./dist/"))))

	// Кастомный NotFound-хендлер.
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[WARN] Не найден URL: %s, метод: %s", r.URL.Path, r.Method)
		acceptHeader := r.Header.Get("Accept")
		if strings.Contains(acceptHeader, "application/json") {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"404 - ресурс не найден"}`, http.StatusNotFound)
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.Error(w, "<h1>404 - Ресурс не найден</h1>", http.StatusNotFound)
		}
	})

	// Применяем глобальный middleware для API, чтобы устанавливать заголовок "Content-Type: application/json".
	router.Use(jsonMiddleware)

	// ----------------------- API-маршруты -----------------------
	api := router.PathPrefix("/api/v1").Subrouter()
	log.Println("[INFO] Регистрация API маршрутов под префиксом /api/v1")

	// Маршруты для запчастей:
	api.HandleFunc("/parts", controllers.GetAllParts).Methods("GET")
	api.HandleFunc("/parts/{id}", controllers.GetPartByID).Methods("GET")
	api.HandleFunc("/parts", controllers.AddPart).Methods("POST")
	api.HandleFunc("/parts/{id}", controllers.UpdatePart).Methods("PUT")
	api.HandleFunc("/parts/{id}", controllers.DeletePart).Methods("DELETE")

	// Маршруты для пользователей:
	api.HandleFunc("/users", controllers.GetAllUsers).Methods("GET")
	api.HandleFunc("/users/{id}", controllers.GetUserByID).Methods("GET")
	api.HandleFunc("/register", controllers.RegisterUserHandler).Methods("POST")
	api.HandleFunc("/login", controllers.AuthorizeUserHandler).Methods("POST")

	// Маршруты для групп:
	api.HandleFunc("/groups", controllers.GetAllGroups).Methods("GET")
	api.HandleFunc("/groups/{id}", controllers.GetGroupByID).Methods("GET")
	api.HandleFunc("/groups", controllers.AddGroup).Methods("POST")
	api.HandleFunc("/groups/{id}", controllers.UpdateGroup).Methods("PUT")
	api.HandleFunc("/groups/{id}", controllers.DeleteGroup).Methods("DELETE")

	// Маршруты для подкатегорий (обновленный ресурс "subcategories"):
	api.HandleFunc("/subcategories", controllers.GetAllSubcategories).Methods("GET")
	api.HandleFunc("/subcategories/{id}", controllers.GetSubcategoryByID).Methods("GET")
	api.HandleFunc("/subcategories", controllers.AddSubcategory).Methods("POST")
	api.HandleFunc("/subcategories/{id}", controllers.UpdateSubcategory).Methods("PUT")
	api.HandleFunc("/subcategories/{id}", controllers.DeleteSubcategory).Methods("DELETE")

	// Новые маршруты для категорий:
	api.HandleFunc("/categories", controllers.GetAllCategories).Methods("GET")
	api.HandleFunc("/categories/{id}", controllers.GetCategoryByID).Methods("GET")
	api.HandleFunc("/categories", controllers.AddCategory).Methods("POST")
	api.HandleFunc("/categories/{id}", controllers.UpdateCategory).Methods("PUT")
	api.HandleFunc("/categories/{id}", controllers.DeleteCategory).Methods("DELETE")

	// ----------------------- HTML-маршруты -----------------------
	pages := router.PathPrefix("").Subrouter()
	pages.Use(htmlMiddleware)
	log.Println("[INFO] Регистрация HTML маршрутов")
	pages.HandleFunc("/", controllers.HomeHandler).Methods("GET")
	pages.HandleFunc("/products", controllers.ProductsHandler).Methods("GET")
	pages.HandleFunc("/register", controllers.RegisterPageHandler).Methods("GET")
	pages.HandleFunc("/login", controllers.LoginPageHandler).Methods("GET")
	pages.HandleFunc("/about", controllers.AboutPageHandler).Methods("GET")
	pages.HandleFunc("/contact", controllers.ContactPageHandler).Methods("GET")
	pages.HandleFunc("/cabinet", controllers.PersonalCabinetHandler).Methods("GET")
	pages.HandleFunc("/logout", controllers.LogoutHandler).Methods("GET")
	pages.HandleFunc("/admin", controllers.AdminPageHandler).Methods("GET")

	// Маршруты для работы с Excel-файлами:
	router.HandleFunc("/admin/import", controllers.AdminImportPartsHandler).Methods("POST")
	router.HandleFunc("/admin/edit_excel", controllers.AdminEditExcelHandler).Methods("GET")
	router.HandleFunc("/admin/save_excel_edits", controllers.AdminSaveExcelEditsHandler).Methods("POST")

	// Обслуживание статических файлов из .well-known
	router.PathPrefix("/.well-known/").Handler(http.FileServer(http.Dir(".")))

	return router
}

// jsonMiddleware устанавливает заголовок "Content-Type: application/json"
// и логирует входящие API-запросы.
func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] API запрос: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// htmlMiddleware устанавливает заголовок "Content-Type: text/html; charset=utf-8"
// и логирует входящие HTML-запросы.
func htmlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] HTML запрос: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}
