package controllers

import (
	"AutoM/config"
	"AutoM/models"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// AdminPageHandler – обработчик страницы админ-панели (/admin).
// Проверяет сессию, затем загружает данные о запчастях и рендерит шаблон admin.html.
func AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем сессию.
	session, err := getSession(w, r)
	if err != nil {
		log.Printf("[ERROR] Не удалось получить сессию: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	// Проверяем, что пользователь – администратор.
	isAdmin, ok := session.Values["is_admin"].(bool)
	if !ok || !isAdmin {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
	}
	// Получаем данные запчастей из базы.
	parts, err := models.GetAllParts(config.DB)
	if err != nil {
		log.Printf("[ERROR] Ошибка получения данных для админ панели: %v", err)
		http.Error(w, "Ошибка получения данных", http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Verified": true,
		"Parts":    parts,
		"Username": session.Values["username"],
	}

	log.Printf("[INFO] Отображение админ-панели для пользователя: %v", session.Values["username"])
	// RenderTemplateCached должен быть реализован в вашем проекте.
	RenderTemplateCached(w, "admin.html", data)
}

// AdminImportPartsHandler – обработчик импорта запчастей из Excel.
// Осуществляет чтение файла, парсит строки (пропуская заголовок),
// создаёт объекты Part и вставляет их в базу.
func AdminImportPartsHandler(w http.ResponseWriter, r *http.Request) {
	// Ограничиваем размер файла 10 МБ.
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Ошибка обработки формы: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем Excel-файл из поля "excel_file".
	file, header, err := r.FormFile("excel_file")
	if err != nil {
		http.Error(w, "Ошибка получения файла: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("[INFO] Загружен Excel-файл: %s, размер: %d байт", header.Filename, header.Size)

	// Открытие файла с использованием excelize.
	f, err := excelize.OpenReader(file)
	if err != nil {
		log.Printf("[ERROR] Ошибка открытия Excel файла: %v", err)
		http.Error(w, "Ошибка чтения Excel файла: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		http.Error(w, "Excel файл не содержит листов", http.StatusBadRequest)
		return
	}
	sheetName := sheetList[0]
	log.Printf("[INFO] Используем лист: %s", sheetName)

	// Чтение строк с выбранного листа.
	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Printf("[ERROR] Ошибка получения строк с листа: %v", err)
		http.Error(w, "Ошибка получения данных из Excel файла: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Предполагается, что первая строка – заголовки.
	importedCount := 0
	for i, row := range rows {
		if i == 0 {
			continue // пропускаем заголовки.
		}
		// Проверяем, чтобы в строке было минимум 5 столбцов:
		// название, описание, цена, subcategory_id, количество.
		if len(row) < 5 {
			log.Printf("[WARN] Строка %d имеет недостаточно столбцов", i+1)
			continue
		}
		name := strings.TrimSpace(row[0])
		description := strings.TrimSpace(row[1])
		price, err := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
		if err != nil {
			log.Printf("[WARN] Строка %d: ошибка преобразования цены: %v", i+1, err)
			continue
		}
		// Преобразование subcategory_id.
		subcategoryID, err := strconv.Atoi(strings.TrimSpace(row[3]))
		if err != nil {
			log.Printf("[WARN] Строка %d: ошибка преобразования SubcategoryID: %v", i+1, err)
			continue
		}
		quantity, err := strconv.Atoi(strings.TrimSpace(row[4]))
		if err != nil {
			log.Printf("[WARN] Строка %d: ошибка преобразования Quantity: %v", i+1, err)
			quantity = 0
		}
		var imageURL *string
		if len(row) > 5 && strings.TrimSpace(row[5]) != "" {
			url := strings.TrimSpace(row[5])
			imageURL = &url
		}
		part := models.Part{
			Name:          name,
			Description:   description,
			Price:         price,
			SubcategoryID: subcategoryID,
			Quantity:      quantity,
			ImageURL:      imageURL,
		}
		if err := models.InsertPart(config.DB, &part); err != nil {
			log.Printf("[ERROR] Строка %d: не удалось вставить товар: %v", i+1, err)
			continue
		}
		importedCount++
	}

	msg := fmt.Sprintf("Импортировано товаров: %d", importedCount)
	log.Printf("[INFO] %s", msg)
	http.Redirect(w, r, "/admin?msg="+msg, http.StatusSeeOther)
}

// AdminEditExcelHandler – обрабатывает запрос на редактирование Excel-файла.
// Открывает Excel по заданному пути и рендерит HTML-шаблон, передавая в него данные.
func AdminEditExcelHandler(w http.ResponseWriter, r *http.Request) {
	const excelPath = "Прайс наш   11.03.2025..xlsx.xlsx" // Укажите тот путь, где находится файл
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		log.Printf("[ERROR] Ошибка открытия Excel файла: %v", err)
		http.Error(w, "Ошибка чтения Excel файла: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		http.Error(w, "Excel файл не содержит ни одного листа", http.StatusBadRequest)
		return
	}
	sheetName := sheetList[0]
	log.Printf("[INFO] Редактирование: используется лист '%s'", sheetName)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Printf("[ERROR] Ошибка получения строк с листа: %v", err)
		http.Error(w, "Ошибка получения строк из Excel файла: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/edit_excel.html")
	if err != nil {
		log.Printf("[ERROR] Ошибка загрузки шаблона: %v", err)
		http.Error(w, "Ошибка загрузки шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, rows); err != nil {
		log.Printf("[ERROR] Ошибка рендеринга шаблона: %v", err)
		http.Error(w, "Ошибка рендеринга шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// AdminSaveExcelEditsHandler – принимает измененные данные из HTML-формы редактирования Excel и выводит их в лог.
func AdminSaveExcelEditsHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("[ERROR] Ошибка обработки формы: %v", err)
		http.Error(w, "Ошибка обработки данных формы: "+err.Error(), http.StatusBadRequest)
		return
	}

	var editedData [][]string
	maxRow, maxCol := 0, 0
	for key := range r.PostForm {
		var row, col int
		_, err := fmt.Sscanf(key, "cell_%d_%d", &row, &col)
		if err == nil {
			if row > maxRow {
				maxRow = row
			}
			if col > maxCol {
				maxCol = col
			}
		}
	}
	editedData = make([][]string, maxRow+1)
	for i := 0; i <= maxRow; i++ {
		editedData[i] = make([]string, maxCol+1)
	}
	for key, val := range r.PostForm {
		var row, col int
		_, err := fmt.Sscanf(key, "cell_%d_%d", &row, &col)
		if err == nil && len(val) > 0 {
			editedData[row][col] = val[0]
		}
	}

	log.Println("[INFO] Итоговые данные после редактирования:")
	for _, row := range editedData {
		log.Println(row)
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
