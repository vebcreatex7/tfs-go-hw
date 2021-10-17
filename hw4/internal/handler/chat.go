package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/tfs-go-hw/hw4/internal/domain"
)

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

type Chat struct {
	c []Message
}

func (chat *Chat) SendMessage(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(UserID).(User)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var parsing domain.Message

	if err = json.Unmarshal(d, &parsing); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chat.c = append(chat.c, Message{Username: u.Username, Message: parsing.Message})
}

func (chat *Chat) ReadMessage(w http.ResponseWriter, r *http.Request) {
	if _, ok := r.Context().Value(UserID).(User); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	text, err := json.Marshal(chat.c)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(text)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
