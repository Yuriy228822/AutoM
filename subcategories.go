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

// GetAllSubcategories – HTTP-обработчик для получения всех записей из таблицы subcategories.
func GetAllSubcategories(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] GetAllSubcategories: начало запроса")
	subcats, err := models.GetAllSubcategories(config.DB)
	if err != nil {
		log.Printf("[ERROR] GetAllSubcategories: ошибка загрузки subcategories: %v", err)
		http.Error(w, "Ошибка загрузки subcategories", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(subcats); err != nil {
		log.Printf("[ERROR] GetAllSubcategories: ошибка кодирования JSON: %v", err)
	}
}

// GetSubcategoryByID – HTTP-обработчик для получения записи из subcategories по subcategory_id.
func GetSubcategoryByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}
	subcat, err := models.GetSubcategoryByID(config.DB, id)
	if err != nil {
		log.Printf("[ERROR] GetSubcategoryByID: %v", err)
		http.Error(w, "Запись не найдена", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subcat)
}

// AddSubcategory – HTTP-обработчик для добавления новой записи в subcategories.
func AddSubcategory(w http.ResponseWriter, r *http.Request) {
	var subcat models.Subcategory
	if err := json.NewDecoder(r.Body).Decode(&subcat); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	// Предполагаем, что клиент передаёт category_id и name.
	if err := models.AddSubcategory(config.DB, subcat.CategoryID, subcat.Name); err != nil {
		log.Printf("[ERROR] AddSubcategory: %v", err)
		http.Error(w, "Ошибка добавления записи", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// UpdateSubcategory – HTTP-обработчик для обновления записи в subcategories.
func UpdateSubcategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}
	var subcat models.Subcategory
	if err := json.NewDecoder(r.Body).Decode(&subcat); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	// Обновляем запись по subcategory_id (id), устанавливая новые значения category_id и name.
	rowsAffected, err := models.UpdateSubcategory(config.DB, id, subcat.CategoryID, subcat.Name)
	if err != nil {
		log.Printf("[ERROR] UpdateSubcategory: %v", err)
		http.Error(w, "Ошибка обновления записи", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Запись не найдена", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DeleteSubcategory – HTTP-обработчик для удаления записи из subcategories.
func DeleteSubcategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}
	rowsAffected, err := models.DeleteSubcategory(config.DB, id)
	if err != nil {
		log.Printf("[ERROR] DeleteSubcategory: %v", err)
		http.Error(w, "Ошибка удаления записи", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Запись не найдена", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}
