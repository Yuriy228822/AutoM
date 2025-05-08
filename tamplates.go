package controllers

// "html/template"
// "log"
// "net/http"
// "path/filepath"
// "sync"

// tplCache хранит распарсенные шаблоны для повторного использования,
// что позволяет избежать повторного парсинга одного и того же файла.
// var (
// 	tplCache   = make(map[string]*template.Template)
// 	cacheMutex sync.RWMutex
// )

// // RenderTemplateCached выполняет кэширование и рендеринг HTML-шаблона.
// // tmplName — имя файла шаблона (например, "home.html"), data — данные, передаваемые в шаблон.
// func RenderTemplateCached(w http.ResponseWriter, tmplName string, data interface{}) {
// 	// Формируем абсолютный путь к файлу шаблона.
// 	tplPath := filepath.Join("templates", tmplName)

// 	// Попытка получить шаблон из кэша.
// 	cacheMutex.RLock()
// 	tpl, ok := tplCache[tmplName]
// 	cacheMutex.RUnlock()

// 	// Если шаблон не найден, парсим его и сохраняем в кэш.
// 	if !ok {
// 		var err error
// 		tpl, err = template.ParseFiles(tplPath)
// 		if err != nil {
// 			log.Printf("[ERROR] Не удалось загрузить шаблон %s: %v", tplPath, err)
// 			http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
// 			return
// 		}
// 		cacheMutex.Lock()
// 		tplCache[tmplName] = tpl
// 		cacheMutex.Unlock()
// 	}

// 	// Выполняем шаблон с переданными данными.
// 	if err := tpl.Execute(w, data); err != nil {
// 		log.Printf("[ERROR] Ошибка исполнения шаблона %s: %v", tplPath, err)
// 		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
// 	}
// }
