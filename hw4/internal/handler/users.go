package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/tfs-go-hw/hw4/internal/domain"
)

const (
	Token      = "Authorization"
	UserID tID = "ID"
)

type tID string

type User struct {
	ID          int
	Name        string
	Username    string
	Password    string
	PrivateChat Chat
}

type Users struct {
	listOfUsers []User
}

func (l *Users) SignUp(w http.ResponseWriter, r *http.Request) {
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var parsing domain.User

	err = json.Unmarshal(d, &parsing)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var u User
	u.ID = len(l.listOfUsers)
	u.Name = parsing.Name
	u.Username = parsing.Username
	u.Password = parsing.Password

	if _, ok := l.FindUser(u); ok {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("User already exists"))
		return
	}
	l.listOfUsers = append(l.listOfUsers, u)

	token, err := CreateToken(u)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Add("token", token)
}

func (l *Users) FindUser(u User) (int, bool) {
	for idx, val := range l.listOfUsers {
		if val.Name == u.Name && val.Username == u.Username && val.Password == u.Password {
			return idx, true
		}
	}
	return -1, false
}

func (l *Users) GetUsers(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(l.listOfUsers)
	_, _ = w.Write(data)
}

func (l *Users) UserInfo(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(UserID).(User)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte("hello, " + fmt.Sprintf("%v", u)))
}

func (l *Users) Auth(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		z := r.Header.Get(Token)
		headerParts := strings.Split(z, " ")
		if len(headerParts) != 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		c := headerParts[1]

		if c == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		u, err := ParceToken(c)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		id, ok := l.FindUser(u)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		idCtx := context.WithValue(r.Context(), UserID, l.listOfUsers[id])
		handler.ServeHTTP(w, r.WithContext(idCtx))
	}

	return http.HandlerFunc(fn)
}

func (l *Users) SendMessage(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(UserID).(User)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if id >= len(l.listOfUsers) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	l.listOfUsers[id].PrivateChat.SendMessage(w, r)
}

func (l *Users) ReadMessage(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(UserID).(User)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	u.PrivateChat.ReadMessage(w, r)
}
