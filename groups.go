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

// GetAllGroups – обработчик для получения списка всех групп в формате JSON.
func GetAllGroups(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] GetAllGroups: начало запроса")

	groups, err := models.GetAllGroups(config.DB)
	if err != nil {
		log.Printf("[ERROR] GetAllGroups: ошибка загрузки групп: %v", err)
		http.Error(w, "Ошибка загрузки групп", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(groups); err != nil {
		log.Printf("[ERROR] GetAllGroups: ошибка кодирования JSON: %v", err)
	} else {
		log.Println("[INFO] GetAllGroups: данные успешно отправлены")
	}
}

// GetGroupByID – обработчик для получения группы по ID.
func GetGroupByID(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] GetGroupByID: начало запроса")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("[ERROR] GetGroupByID: неверный ID, ошибка преобразования: %v", err)
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	group, err := models.GetGroupByID(config.DB, id)
	if err != nil {
		log.Printf("[ERROR] GetGroupByID: ошибка получения группы: %v", err)
		http.Error(w, "Группа не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(group); err != nil {
		log.Printf("[ERROR] GetGroupByID: ошибка кодирования JSON: %v", err)
	} else {
		log.Printf("[INFO] GetGroupByID: группа с ID %d успешно отправлена", id)
	}
}

// AddGroup – обработчик для добавления новой группы.
func AddGroup(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] AddGroup: начало запроса")

	var group models.Group
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		log.Printf("[ERROR] AddGroup: ошибка декодирования JSON: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Вставка осуществляется только по имени, остальные поля генерируются базой.
	if err := models.AddGroup(config.DB, group.Name); err != nil {
		log.Printf("[ERROR] AddGroup: ошибка добавления группы: %v", err)
		http.Error(w, "Ошибка добавления группы", http.StatusInternalServerError)
		return
	}

	log.Println("[INFO] AddGroup: группа успешно добавлена")
	w.WriteHeader(http.StatusCreated)
}

// UpdateGroup – обработчик для обновления группы.
func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] UpdateGroup: начало запроса")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("[ERROR] UpdateGroup: неверный ID, ошибка преобразования: %v", err)
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	var group models.Group
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		log.Printf("[ERROR] UpdateGroup: ошибка декодирования JSON: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	rowsAffected, err := models.UpdateGroup(config.DB, id, group.Name)
	if err != nil {
		log.Printf("[ERROR] UpdateGroup: ошибка обновления группы: %v", err)
		http.Error(w, "Ошибка обновления группы", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("[ERROR] UpdateGroup: группа с ID %d не найдена", id)
		http.Error(w, "Группа не найдена", http.StatusNotFound)
		return
	}

	log.Printf("[INFO] UpdateGroup: группа с ID %d успешно обновлена", id)
	w.WriteHeader(http.StatusOK)
}

// DeleteGroup – обработчик для удаления группы.
func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] DeleteGroup: начало запроса")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("[ERROR] DeleteGroup: неверный ID, ошибка преобразования: %v", err)
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	rowsAffected, err := models.DeleteGroup(config.DB, id)
	if err != nil {
		log.Printf("[ERROR] DeleteGroup: ошибка удаления группы: %v", err)
		http.Error(w, "Ошибка удаления группы", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("[ERROR] DeleteGroup: группа с ID %d не найдена", id)
		http.Error(w, "Группа не найдена", http.StatusNotFound)
		return
	}

	log.Printf("[INFO] DeleteGroup: группа с ID %d успешно удалена", id)
	w.WriteHeader(http.StatusOK)
}
