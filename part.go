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

// GetAllParts – получение всех запчастей (API, JSON-вывод)
func GetAllParts(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Начало запроса GetAllParts")
	// Обновляем запрос: выбираем part_id как id и subcategory_id вместо category_id
	query := "SELECT part_id AS id, name, description, price, image_url, subcategory_id, quantity, create_at, update_at FROM parts"
	rows, err := config.DB.Query(query)
	if err != nil {
		log.Printf("[ERROR] GetAllParts: ошибка запроса: %v", err)
		http.Error(w, "Ошибка загрузки запчастей", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var parts []models.Part
	i := 0
	for rows.Next() {
		i++
		var part models.Part
		if err := rows.Scan(&part.ID, &part.Name, &part.Description, &part.Price, &part.ImageURL, &part.SubcategoryID, &part.Quantity, &part.CreatedAt, &part.UpdatedAt); err != nil {
			log.Printf("[ERROR] GetAllParts: ошибка сканирования строки %d: %v", i, err)
			http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
			return
		}
		parts = append(parts, part)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] GetAllParts: ошибка итерации по строкам: %v", err)
		http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] GetAllParts: успешно получено %d записей", len(parts))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(parts); err != nil {
		log.Printf("[ERROR] GetAllParts: ошибка кодирования JSON: %v", err)
	}
}

// GetPartByID – получение запчасти по ID (API)
func GetPartByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Начало запроса GetPartByID")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("[ERROR] GetPartByID: неверное значение ID: %v", err)
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}
	// Используем part_id AS id и subcategory_id вместо category_id
	query := "SELECT part_id AS id, name, description, price, image_url, subcategory_id, quantity, create_at, update_at FROM parts WHERE part_id = ?"
	row := config.DB.QueryRow(query, id)
	var part models.Part
	if err := row.Scan(&part.ID, &part.Name, &part.Description, &part.Price, &part.ImageURL, &part.SubcategoryID, &part.Quantity, &part.CreatedAt, &part.UpdatedAt); err != nil {
		log.Printf("[ERROR] GetPartByID: запчасть с ID %d не найдена: %v", id, err)
		http.Error(w, "Запчасть не найдена", http.StatusNotFound)
		return
	}
	log.Printf("[INFO] GetPartByID: успешно получена запчасть с ID %d", id)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(part); err != nil {
		log.Printf("[ERROR] GetPartByID: ошибка кодирования JSON: %v", err)
	}
}

// AddPart – добавление новой запчасти (API)
func AddPart(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Начало запроса AddPart")
	var part models.Part
	if err := json.NewDecoder(r.Body).Decode(&part); err != nil {
		log.Printf("[ERROR] AddPart: ошибка декодирования данных: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	// Запрос обновлён: subcategory_id вместо category_id, используются NOW() для временных меток
	query := "INSERT INTO parts (name, description, price, image_url, subcategory_id, quantity, create_at, update_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())"
	_, err := config.DB.Exec(query, part.Name, part.Description, part.Price, part.ImageURL, part.SubcategoryID, part.Quantity)
	if err != nil {
		log.Printf("[ERROR] AddPart: ошибка выполнения запроса: %v", err)
		http.Error(w, "Ошибка добавления запчасти", http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] AddPart: запчасть успешно добавлена: %s", part.Name)
	w.WriteHeader(http.StatusCreated)
}

// UpdatePart – обновление запчасти (API)
func UpdatePart(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Начало запроса UpdatePart")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("[ERROR] UpdatePart: неверное значение ID: %v", err)
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	var part models.Part
	if err := json.NewDecoder(r.Body).Decode(&part); err != nil {
		log.Printf("[ERROR] UpdatePart: ошибка декодирования данных: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	// Обновляем запрос: subcategory_id вместо category_id; используется WHERE part_id = ?
	query := "UPDATE parts SET subcategory_id = ?, name = ?, price = ?, quantity = ?, description = ? WHERE part_id = ?"
	res, err := config.DB.Exec(query, part.SubcategoryID, part.Name, part.Price, part.Quantity, part.Description, id)
	if err != nil {
		log.Printf("[ERROR] UpdatePart: ошибка выполнения запроса: %v", err)
		http.Error(w, "Ошибка обновления запчасти", http.StatusInternalServerError)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] UpdatePart: ошибка получения количества затронутых строк: %v", err)
		http.Error(w, "Ошибка обновления запчасти", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Printf("[WARN] UpdatePart: запчасть с ID %d не найдена", id)
		http.Error(w, "Запчасть не найдена", http.StatusNotFound)
		return
	}
	log.Printf("[INFO] UpdatePart: запчасть с ID %d успешно обновлена", id)
	w.WriteHeader(http.StatusOK)
}

// DeletePart – удаление запчасти (API)
func DeletePart(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] Начало запроса DeletePart")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("[ERROR] DeletePart: неверное значение ID: %v", err)
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}
	// Обновляем запрос: удаляем по part_id вместо id
	query := "DELETE FROM parts WHERE part_id = ?"
	res, err := config.DB.Exec(query, id)
	if err != nil {
		log.Printf("[ERROR] DeletePart: ошибка выполнения запроса: %v", err)
		http.Error(w, "Ошибка удаления запчасти", http.StatusInternalServerError)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] DeletePart: ошибка получения количества затронутых строк: %v", err)
		http.Error(w, "Ошибка удаления запчасти", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		log.Printf("[WARN] DeletePart: запчасть с ID %d не найдена", id)
		http.Error(w, "Запчасть не найдена", http.StatusNotFound)
		return
	}
	log.Printf("[INFO] DeletePart: запчасть с ID %d успешно удалена", id)
	w.WriteHeader(http.StatusOK)
}
