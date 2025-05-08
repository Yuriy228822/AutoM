package controllers

import (
	"AutoM/config"
	"AutoM/models"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// HomePageData – структура для передачи данных в шаблон главной страницы.
type HomePageData struct {
	Username         string
	IsAdmin          bool
	Parts            []models.Part
	Categories       []models.Category
	SelectedCategory string
}

var (
	homeTpl *template.Template
	tplOnce sync.Once
)

// loadHomeTemplate загружает шаблон home.html один раз при первом обращении.
func loadHomeTemplate() {
	tplPath := filepath.Join("templates", "home.html")
	var err error
	homeTpl, err = template.ParseFiles(tplPath)
	if err != nil {
		log.Fatalf("[FATAL] Ошибка загрузки шаблона %s: %v", tplPath, err)
	}
}

// HomeHandler обрабатывает запрос на главную страницу и осуществляет выборку товаров.
// Если передан GET-параметр "category", происходит выборка товаров по данной категории;
// если параметр не указан – загружаются все товары.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Загружаем шаблон один раз.
	tplOnce.Do(loadHomeTemplate)

	// Получаем выбранную категорию из параметра запроса.
	selectedCategory := strings.TrimSpace(r.URL.Query().Get("category"))
	var parts []models.Part
	var err error

	// Если категория выбрана, пробуем преобразовать её к числовому ID.
	if selectedCategory != "" {
		if subcategoryID, convErr := strconv.Atoi(selectedCategory); convErr == nil {
			parts, err = models.GetPartsBySubcategory(config.DB, subcategoryID)
		} else {
			// Если не число - выбираем товары по имени категории.
			parts, err = models.GetPartsByCategoryName(config.DB, selectedCategory)
		}
		if err != nil {
			log.Printf("[ERROR] Ошибка получения товаров по категории (%s): %v", selectedCategory, err)
			http.Error(w, "Ошибка получения товаров", http.StatusInternalServerError)
			return
		}
	} else {
		// Если категория не выбрана, загружаем все товары.
		parts, err = models.GetAllParts(config.DB)
		if err != nil {
			log.Printf("[ERROR] Ошибка получения товаров: %v", err)
			http.Error(w, "Ошибка получения товаров", http.StatusInternalServerError)
			return
		}
	}

	// Получаем все категории для формирования меню фильтрации.
	categories, err := models.GetAllCategories(config.DB)
	if err != nil {
		log.Printf("[ERROR] Ошибка получения категорий: %v", err)
		// Если возникла ошибка, продолжаем с пустым списком категорий.
		categories = []models.Category{}
	}

	// Извлекаем из сессии данные пользователя.
	session, _ := config.Store.Get(r, "session")
	var username string
	var isAdmin bool
	if val, ok := session.Values["username"].(string); ok {
		username = strings.TrimSpace(val)
	}
	if adminVal, ok := session.Values["is_admin"].(bool); ok {
		isAdmin = adminVal
	} else {
		log.Println("[DEBUG] Значение is_admin отсутствует или имеет неверный тип")
	}

	// Формируем данные для шаблона.
	data := HomePageData{
		Username:         username,
		IsAdmin:          isAdmin,
		Parts:            parts,
		Categories:       categories,
		SelectedCategory: selectedCategory,
	}

	// Рендерим шаблон и передаём данные.
	if err := homeTpl.Execute(w, data); err != nil {
		log.Printf("[ERROR] Ошибка исполнения шаблона главной страницы: %v", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}
