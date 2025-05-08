package controllers

import (
	"net/http"
)

// ProductsHandler – обработчик страницы со списком товаров.
// В реальном приложении данные можно получить из базы данных и передать в шаблон через data.
func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	RenderTemplateCached(w, "products.html", nil)
}

// RegisterPageHandler – обработчик страницы регистрации пользователя.
func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	RenderTemplateCached(w, "register.html", nil)
}

// LoginPageHandler – обработчик страницы авторизации (входа).
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	RenderTemplateCached(w, "login.html", nil)
}

// AboutPageHandler – обработчик страницы "О нас".
func AboutPageHandler(w http.ResponseWriter, r *http.Request) {
	RenderTemplateCached(w, "about.html", nil)
}

// ContactPageHandler – обработчик страницы "Контакты".
func ContactPageHandler(w http.ResponseWriter, r *http.Request) {
	RenderTemplateCached(w, "contact.html", nil)
}
