package main

import (
	"net/http"
	"encoding/json"
)

type user struct {
	User struct {
		Username	string 	`json:"username"`
		Email		string	`json:"email"`
		CreatedAt	string	`json:"createdAt"`
		UpdatedAt	string	`json:"updatedAt"`
		Password	string	`json:"password"`
		Token		string 	`json:"token"`
	} `json:"user"`
}

func (u *user) encodeUser (r *http.Request) (err error) {
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&u)
	return
}

func (u *user) decodeUser () (jsonData []byte) {
	jsonData, _ = json.Marshal(u)
	return
}
