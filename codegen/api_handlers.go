package main

import (
	"net/http"
	"encoding/json"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/schema"
	"fmt"
	)

type resp map[string]interface{}

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
switch r.URL.Path {
	case "/user/profile":
		h.handlerProfile(w, r)
	case "/user/create":
		h.handlerCreate(w, r)
			}
}

func (h *MyApi) handlerProfile (w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := ProfileParams{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err := decoder.Decode(params, r.URL.Query())
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal", 500)
		return
	}
	_, err = govalidator.ValidateStruct(params)
	if err != nil {
		if allErrs, ok := err.(govalidator.Errors); ok {
			for _, fld := range allErrs.Errors() {
				data := []byte(fmt.Sprintf("field: %#v\n\n", fld))
				w.Write(data)
			}
		}
	}
	res, err := h.Profile(ctx, params)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	jsonResp, _ := json.Marshal(res)
	w.Write(jsonResp)
	return
}

func (h *MyApi) handlerCreate (w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := CreateParams{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err := decoder.Decode(params, r.URL.Query())
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal", 500)
		return
	}
	_, err = govalidator.ValidateStruct(params)
	if err != nil {
		if allErrs, ok := err.(govalidator.Errors); ok {
			for _, fld := range allErrs.Errors() {
				data := []byte(fmt.Sprintf("field: %#v\n\n", fld))
				w.Write(data)
			}
		}
	}
	res, err := h.Create(ctx, params)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	jsonResp, _ := json.Marshal(res)
	w.Write(jsonResp)
	return
}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
switch r.URL.Path {
	case "/user/create":
		h.handlerCreate(w, r)
			}
}

func (h *OtherApi) handlerCreate (w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := OtherCreateParams{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	err := decoder.Decode(params, r.URL.Query())
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal", 500)
		return
	}
	_, err = govalidator.ValidateStruct(params)
	if err != nil {
		if allErrs, ok := err.(govalidator.Errors); ok {
			for _, fld := range allErrs.Errors() {
				data := []byte(fmt.Sprintf("field: %#v\n\n", fld))
				w.Write(data)
			}
		}
	}
	res, err := h.Create(ctx, params)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	jsonResp, _ := json.Marshal(res)
	w.Write(jsonResp)
	return
}
