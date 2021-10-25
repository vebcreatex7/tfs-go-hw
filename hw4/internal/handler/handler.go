package handler

import (
	"sync"

	"github.com/go-chi/chi/v5"
)

type Handler struct{}

func (h *Handler) InitRoutes() *chi.Mux {
	Users := &Users{}
	Chat := &Chat{mu: &sync.Mutex{}}

	root := chi.NewRouter()
	root.Post("/sign-up", Users.SignUp)

	u := chi.NewRouter()
	u.Use(Users.Auth)
	u.Post("/{id}/message", Users.SendMessage)
	u.Get("/me/message", Users.ReadMessage)
	root.Mount("/user", u)

	m := chi.NewRouter()
	m.Use(Users.Auth)
	m.Post("/send", Chat.SendMessage)
	m.Get("/read", Chat.ReadMessage)
	root.Mount("/message", m)

	root.Get("/list", Users.GetUsers)

	return root
}
