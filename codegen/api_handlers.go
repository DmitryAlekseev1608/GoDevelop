package main

	import (
		"net/http"
		"encoding/json"

	)
type resp map[string]interface{}

// MyApi
func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// var auth = r.Header.Get("X-Auth")
	// if auth != "100500" {
	// 	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	// 	return
	// }
switch r.URL.Path {
	case "/user/profile":
		var result = CR{
				"id":        42,
				"login":     "rvasily",
				"full_name": "Vasily Romanov",
				"status":    20,
			}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		jsonResp, _ := json.Marshal(result)
		w.Write(jsonResp)
		return

}

// func (h *MyApi) wrapperDoSomeJob() {
// 	res, err := h.DoSomeJob(ctx, params)
}

// OtherApi
func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// var auth = r.Header.Get("X-Auth")
	// if auth != "100500" {
	// 	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	// 	return
	// }
switch r.URL.Path {
	case "/user/profile":
		var result = resp{
				"id":        42,
				"login":     "rvasily",
				"full_name": "Vasily Romanov",
				"status":    20,
			}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		jsonResp, _ := json.Marshal(result)
		w.Write(jsonResp)
		return

}

// func (h *OtherApi) wrapperDoSomeJob() {
// 	res, err := h.DoSomeJob(ctx, params)
}
