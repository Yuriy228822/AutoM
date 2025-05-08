package models

import (
	"database/sql"
	"time"
)

// Subcategory описывает запись из таблицы subcategories,
// где каждая подкатегория определяется своим subcategory_id и принадлежит категории (category_id).
type Subcategory struct {
	SubcategoryID int       `json:"subcategory_id"` // первичный ключ
	CategoryID    int       `json:"category_id"`    // идентификатор родительской категории
	Name          string    `json:"name"`           // название подкатегории
	CreateAt      time.Time `json:"create_at"`      // время создания
	UpdateAt      time.Time `json:"update_at"`      // время обновления
}

// GetAllSubcategories возвращает список всех подкатегорий.
func GetAllSubcategories(db *sql.DB) ([]Subcategory, error) {
	query := "SELECT subcategory_id, category_id, name, create_at, update_at FROM subcategories"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []Subcategory
	for rows.Next() {
		var s Subcategory
		if err := rows.Scan(&s.SubcategoryID, &s.CategoryID, &s.Name, &s.CreateAt, &s.UpdateAt); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

// GetSubcategoryByID возвращает подкатегорию по её subcategory_id.
func GetSubcategoryByID(db *sql.DB, id int) (Subcategory, error) {
	var s Subcategory
	query := "SELECT subcategory_id, category_id, name, create_at, update_at FROM subcategories WHERE subcategory_id = ?"
	err := db.QueryRow(query, id).Scan(&s.SubcategoryID, &s.CategoryID, &s.Name, &s.CreateAt, &s.UpdateAt)
	if err != nil {
		return Subcategory{}, err
	}
	return s, nil
}

// AddSubcategory добавляет новую подкатегорию в базу данных.
// Временные метки создаются с помощью NOW().
func AddSubcategory(db *sql.DB, categoryID int, name string) error {
	query := "INSERT INTO subcategories (category_id, name, create_at, update_at) VALUES (?, ?, NOW(), NOW())"
	_, err := db.Exec(query, categoryID, name)
	return err
}

// UpdateSubcategory обновляет информацию о подкатегории по её subcategory_id.
func UpdateSubcategory(db *sql.DB, subcategoryID int, categoryID int, name string) (int64, error) {
	query := "UPDATE subcategories SET category_id = ?, name = ?, update_at = NOW() WHERE subcategory_id = ?"
	result, err := db.Exec(query, categoryID, name, subcategoryID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteSubcategory удаляет подкатегорию по её subcategory_id.
func DeleteSubcategory(db *sql.DB, subcategoryID int) (int64, error) {
	query := "DELETE FROM subcategories WHERE subcategory_id = ?"
	result, err := db.Exec(query, subcategoryID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
