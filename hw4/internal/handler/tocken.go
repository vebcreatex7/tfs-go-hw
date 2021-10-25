package handler

import (
	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	ID       int
	Name     string
	Username string
	Password string
	jwt.StandardClaims
}

var jwtKey = []byte("my_secret_key")

func CreateToken(u User) (string, error) {
	claims := &Claims{
		Username: u.Username,
		Name:     u.Name,
		Password: u.Password,
		ID:       u.ID,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

func ParceToken(tknStr string) (User, error) {
	claim := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claim, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !tkn.Valid {
		return User{}, err
	}

	return User{
		ID:       claim.ID,
		Name:     claim.Name,
		Username: claim.Username,
		Password: claim.Password,
	}, nil
}
