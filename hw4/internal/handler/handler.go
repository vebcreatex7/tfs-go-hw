package handler

import "github.com/go-chi/chi/v5"

type Handler struct {
}

func (h *Handler) InitRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/users", func(r chi.Router) {
		r.Post("/sign-up", SignUp)                     // Регистрация
		r.Post("/{id}/messages", SendPrivateMessage)   // Отправить личное сообщение пользователю id
		r.Get("/me/{id}/messages", VievPrivateMessage) // Получить сообщение от пользователя id

	})

	r.Route("/chat", func(r chi.Router) {
		r.Post("/send", SendMessage) // Отправить сообщение в общий чат
		r.Get("/view", VievMessage)  // Получить сообщение из общего чата
	})

	return r
}
