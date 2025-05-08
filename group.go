package models

import (
	"database/sql"
	"log"
	"time"
)

// Group описывает запись из таблицы групп.
// Таблица имеет следующие столбцы:
// group_id (PK, auto_increment), name, create_at, update_at.
type Group struct {
	ID       int       `json:"id"`        // group_id – первичный ключ
	Name     string    `json:"name"`      // название группы
	CreateAt time.Time `json:"create_at"` // время создания
	UpdateAt time.Time `json:"update_at"` // время последнего обновления
}

// GetAllGroups возвращает список всех групп из базы данных.
func GetAllGroups(db *sql.DB) ([]Group, error) {
	log.Println("[INFO] GetAllGroups: начало запроса")
	// Экранирование имени таблицы для избежания проблем с зарезервированными словами
	query := "SELECT group_id AS id, name, create_at, update_at FROM `groups_main`"
	log.Printf("[INFO] GetAllGroups: выполняется запрос: %s", query)

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("[ERROR] GetAllGroups: ошибка выполнения запроса: %v", err)
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("[ERROR] GetAllGroups: ошибка закрытия rows: %v", err)
		} else {
			log.Println("[INFO] GetAllGroups: rows успешно закрыты")
		}
	}()

	var groups []Group
	rowNum := 0
	for rows.Next() {
		rowNum++
		var grp Group
		if err := rows.Scan(&grp.ID, &grp.Name, &grp.CreateAt, &grp.UpdateAt); err != nil {
			log.Printf("[ERROR] GetAllGroups: ошибка сканирования строки %d: %v", rowNum, err)
			return nil, err
		}
		log.Printf("[INFO] GetAllGroups: строка %d успешно обработана: %+v", rowNum, grp)
		groups = append(groups, grp)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] GetAllGroups: ошибка итерации по строкам: %v", err)
		return nil, err
	}
	log.Printf("[INFO] GetAllGroups: успешно получено %d групп", len(groups))
	return groups, nil
}

// GetGroupByID возвращает группу по её ID.
func GetGroupByID(db *sql.DB, id int) (Group, error) {
	log.Printf("[INFO] GetGroupByID: запрос группы с ID %d", id)
	var grp Group
	query := "SELECT group_id AS id, name, create_at, update_at FROM `groups_main` WHERE group_id = ?"
	log.Printf("[INFO] GetGroupByID: выполняется запрос: %s", query)
	err := db.QueryRow(query, id).Scan(&grp.ID, &grp.Name, &grp.CreateAt, &grp.UpdateAt)
	if err != nil {
		log.Printf("[ERROR] GetGroupByID: ошибка получения группы с ID %d: %v", id, err)
		return Group{}, err
	}
	log.Printf("[INFO] GetGroupByID: получена группа: %+v", grp)
	return grp, nil
}

// AddGroup добавляет новую группу в базу данных.
func AddGroup(db *sql.DB, name string) error {
	log.Printf("[INFO] AddGroup: добавление новой группы: '%s'", name)
	// Вставка производится только по имени; временные метки генерируются базой
	query := "INSERT INTO `groups_main` (name, create_at, update_at) VALUES (?, NOW(), NOW())"
	log.Printf("[INFO] AddGroup: выполняется запрос: %s", query)
	_, err := db.Exec(query, name)
	if err != nil {
		log.Printf("[ERROR] AddGroup: ошибка выполнения запроса: %v", err)
		return err
	}
	log.Printf("[INFO] AddGroup: группа '%s' успешно добавлена", name)
	return nil
}

// UpdateGroup обновляет данные группы по её ID.
func UpdateGroup(db *sql.DB, id int, name string) (int64, error) {
	log.Printf("[INFO] UpdateGroup: обновление группы с ID %d", id)
	query := "UPDATE `groups_main` SET name = ?, update_at = NOW() WHERE group_id = ?"
	log.Printf("[INFO] UpdateGroup: выполняется запрос: %s", query)
	result, err := db.Exec(query, name, id)
	if err != nil {
		log.Printf("[ERROR] UpdateGroup: ошибка выполнения запроса: %v", err)
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] UpdateGroup: ошибка получения количества затронутых строк: %v", err)
		return 0, err
	}
	if rowsAffected == 0 {
		log.Printf("[WARN] UpdateGroup: группа с ID %d не найдена", id)
	} else {
		log.Printf("[INFO] UpdateGroup: группа с ID %d успешно обновлена", id)
	}
	return rowsAffected, nil
}

// DeleteGroup удаляет группу из базы данных по её ID.
func DeleteGroup(db *sql.DB, id int) (int64, error) {
	log.Printf("[INFO] DeleteGroup: удаление группы с ID %d", id)
	query := "DELETE FROM `groups_main` WHERE group_id = ?"
	log.Printf("[INFO] DeleteGroup: выполняется запрос: %s", query)
	result, err := db.Exec(query, id)
	if err != nil {
		log.Printf("[ERROR] DeleteGroup: ошибка выполнения запроса: %v", err)
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] DeleteGroup: ошибка получения количества затронутых строк: %v", err)
		return 0, err
	}
	if rowsAffected == 0 {
		log.Printf("[WARN] DeleteGroup: группа с ID %d не найдена", id)
	} else {
		log.Printf("[INFO] DeleteGroup: группа с ID %d успешно удалена", id)
	}
	return rowsAffected, nil
}
