package handler

import "net/http"

func SignUp(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("sign-up"))
}

func SendPrivateMessage(w http.ResponseWriter, r *http.Request) {

}

func VievPrivateMessage(w http.ResponseWriter, r *http.Request) {

}
