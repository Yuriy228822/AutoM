package models

import (
	"database/sql"
	"log"
	"time"
)

// Category описывает категорию системы.
// Структура отражает таблицу, где присутствуют следующие поля:
// category_id, group_id, name, create_at, update_at и description.
type Category struct {
	ID          int       `json:"id"`          // category_id – первичный ключ
	GroupID     int       `json:"group_id"`    // group_id (идентификатор группы)
	Name        string    `json:"name"`        // название категории
	CreateAt    time.Time `json:"create_at"`   // время создания
	UpdateAt    time.Time `json:"update_at"`   // время последнего обновления
	Description string    `json:"description"` // описание категории
}

// GetAllCategories возвращает список всех категорий из базы данных
// и записывает подробные логи работы.
func GetAllCategories(db *sql.DB) ([]Category, error) {
	log.Printf("[INFO] GetAllCategories: начало выполнения функции")
	// Экранирование имени таблицы обратными кавычками
	query := "SELECT category_id AS id, group_id, name, create_at, update_at, description FROM `categories`"
	log.Printf("[INFO] GetAllCategories: выполнение запроса: %s", query)
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("[ERROR] GetAllCategories: ошибка выполнения запроса: %v", err)
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("[ERROR] GetAllCategories: ошибка закрытия rows: %v", err)
		} else {
			log.Printf("[INFO] GetAllCategories: rows успешно закрыты")
		}
	}()

	var categories []Category
	rowNum := 0
	for rows.Next() {
		rowNum++
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.GroupID, &cat.Name, &cat.CreateAt, &cat.UpdateAt, &cat.Description); err != nil {
			log.Printf("[ERROR] GetAllCategories: ошибка сканирования строки %d: %v", rowNum, err)
			return nil, err
		}
		log.Printf("[INFO] GetAllCategories: строка %d успешно обработана: %+v", rowNum, cat)
		categories = append(categories, cat)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] GetAllCategories: ошибка итерации по строкам: %v", err)
		return nil, err
	}
	log.Printf("[INFO] GetAllCategories: успешно получено %d категорий", len(categories))
	return categories, nil
}

// GetCategoryByID возвращает категорию по её ID.
func GetCategoryByID(db *sql.DB, id int) (Category, error) {
	log.Printf("[INFO] GetCategoryByID: запрос категории с ID %d", id)
	var cat Category
	query := "SELECT category_id AS id, group_id, name, create_at, update_at, description FROM `categories` WHERE category_id = ?"
	log.Printf("[INFO] GetCategoryByID: выполняется запрос: %s", query)
	err := db.QueryRow(query, id).Scan(&cat.ID, &cat.GroupID, &cat.Name, &cat.CreateAt, &cat.UpdateAt, &cat.Description)
	if err != nil {
		log.Printf("[ERROR] GetCategoryByID: ошибка получения категории с ID %d: %v", id, err)
		return Category{}, err
	}
	log.Printf("[INFO] GetCategoryByID: получена категория: %+v", cat)
	return cat, nil
}

// AddCategory добавляет новую категорию в базу данных.
func AddCategory(db *sql.DB, groupID int, name, description string) error {
	log.Printf("[INFO] AddCategory: добавление категории '%s'", name)
	// Вставка производится по group_id, name и description;
	// временные метки генерируются базой через NOW().
	query := "INSERT INTO `categories` (group_id, name, create_at, update_at, description) VALUES (?, ?, NOW(), NOW(), ?)"
	log.Printf("[INFO] AddCategory: выполняется запрос: %s", query)
	_, err := db.Exec(query, groupID, name, description)
	if err != nil {
		log.Printf("[ERROR] AddCategory: ошибка выполнения запроса: %v", err)
		return err
	}
	log.Printf("[INFO] AddCategory: категория '%s' успешно добавлена", name)
	return nil
}

// UpdateCategory обновляет существующую категорию по её ID.
func UpdateCategory(db *sql.DB, id, groupID int, name, description string) (int64, error) {
	log.Printf("[INFO] UpdateCategory: обновление категории с ID %d", id)
	query := "UPDATE `categories` SET group_id = ?, name = ?, update_at = NOW(), description = ? WHERE category_id = ?"
	log.Printf("[INFO] UpdateCategory: выполняется запрос: %s", query)
	result, err := db.Exec(query, groupID, name, description, id)
	if err != nil {
		log.Printf("[ERROR] UpdateCategory: ошибка выполнения запроса: %v", err)
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] UpdateCategory: ошибка получения количества затронутых строк: %v", err)
		return 0, err
	}
	if rowsAffected == 0 {
		log.Printf("[WARN] UpdateCategory: категория с ID %d не найдена", id)
	} else {
		log.Printf("[INFO] UpdateCategory: категория с ID %d успешно обновлена", id)
	}
	return rowsAffected, nil
}

// DeleteCategory удаляет категорию из базы данных по её ID.
func DeleteCategory(db *sql.DB, id int) (int64, error) {
	log.Printf("[INFO] DeleteCategory: удаление категории с ID %d", id)
	query := "DELETE FROM `categories` WHERE category_id = ?"
	log.Printf("[INFO] DeleteCategory: выполняется запрос: %s", query)
	result, err := db.Exec(query, id)
	if err != nil {
		log.Printf("[ERROR] DeleteCategory: ошибка выполнения запроса: %v", err)
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] DeleteCategory: ошибка получения количества затронутых строк: %v", err)
		return 0, err
	}
	if rowsAffected == 0 {
		log.Printf("[WARN] DeleteCategory: категория с ID %d не найдена", id)
	} else {
		log.Printf("[INFO] DeleteCategory: категория с ID %d успешно удалена", id)
	}
	return rowsAffected, nil
}
