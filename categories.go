package controllers

import (
	"AutoM/config"
	"AutoM/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GetAllCategories – обработчик для получения всех категорий (JSON-вывод).
func GetAllCategories(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] GetAllCategories: начало запроса")
	categories, err := models.GetAllCategories(config.DB)
	if err != nil {
		log.Printf("[ERROR] GetAllCategories: ошибка загрузки категорий: %v", err)
		http.Error(w, "Ошибка загрузки категорий", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(categories); err != nil {
		log.Printf("[ERROR] GetAllCategories: ошибка кодирования JSON: %v", err)
	} else {
		log.Println("[INFO] GetAllCategories: данные успешно отправлены")
	}
}

// GetCategoryByID – обработчик для получения категории по её ID.
func GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] GetCategoryByID: начало запроса")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("[ERROR] GetCategoryByID: неверный формат ID: %v", err)
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	category, err := models.GetCategoryByID(config.DB, id)
	if err != nil {
		log.Printf("[ERROR] GetCategoryByID: ошибка получения категории: %v", err)
		http.Error(w, "Категория не найдена", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(category); err != nil {
		log.Printf("[ERROR] GetCategoryByID: ошибка кодирования JSON: %v", err)
		return
	}
	log.Printf("[INFO] GetCategoryByID: категория с ID %d успешно отправлена", id)
}

// AddCategory – обработчик для добавления новой категории.
// Ожидается JSON с полями group_id, name и description.
func AddCategory(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] AddCategory: начало запроса")
	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		log.Printf("[ERROR] AddCategory: ошибка декодирования JSON: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	log.Printf("[INFO] AddCategory: получены данные категории: %+v", category)
	// Добавление категории; функция модели получает group_id, name и description.
	if err := models.AddCategory(config.DB, category.GroupID, category.Name, category.Description); err != nil {
		log.Printf("[ERROR] AddCategory: ошибка добавления категории: %v", err)
		http.Error(w, "Ошибка добавления категории", http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] AddCategory: категория '%s' успешно добавлена", category.Name)
	w.WriteHeader(http.StatusCreated)
}

// UpdateCategory – обработчик для обновления категории.
// Ожидается JSON с обновленными данными: group_id, name и description.
func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] UpdateCategory: начало запроса")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("[ERROR] UpdateCategory: неверный формат ID: %v", err)
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}
	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		log.Printf("[ERROR] UpdateCategory: ошибка декодирования JSON: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	log.Printf("[INFO] UpdateCategory: получены данные обновления: %+v", category)
	rowsAffected, err := models.UpdateCategory(config.DB, id, category.GroupID, category.Name, category.Description)
	if err != nil {
		log.Printf("[ERROR] UpdateCategory: ошибка обновления категории: %v", err)
		http.Error(w, "Ошибка обновления категории", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Printf("[WARN] UpdateCategory: категория с ID %d не найдена", id)
		http.Error(w, "Категория не найдена", http.StatusNotFound)
		return
	}
	log.Printf("[INFO] UpdateCategory: категория с ID %d успешно обновлена", id)
	w.WriteHeader(http.StatusOK)
}

// DeleteCategory – обработчик для удаления категории по её ID.
func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] DeleteCategory: начало запроса")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("[ERROR] DeleteCategory: неверный формат ID: %v", err)
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}
	rowsAffected, err := models.DeleteCategory(config.DB, id)
	if err != nil {
		log.Printf("[ERROR] DeleteCategory: ошибка удаления категории: %v", err)
		http.Error(w, "Ошибка удаления категории", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Printf("[WARN] DeleteCategory: категория с ID %d не найдена", id)
		http.Error(w, "Категория не найдена", http.StatusNotFound)
		return
	}
	log.Printf("[INFO] DeleteCategory: категория с ID %d успешно удалена", id)
	w.WriteHeader(http.StatusOK)
}
