	// var auth = r.Header.Get("X-Auth")
	// if auth != "100500" {
	// 	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	// 	return
	// }


		}
}
`))
	methodTpl = template.Must(template.New("methodTpl").Parse(`
// {{.StructName}}
// {{.MethodName}}
func (h *{{.FieldName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
switch r.URL.Path {
	case "/user/profile":


	return
	}

func (h *{{.FieldName}}) wrapperDoSomeJob() {
	res, err := h.DoSomeJob(ctx, params)
	var result = resp{
		"error": "",
		"response": resp{
			"id":        42,
			"login":     "rvasily",
			"full_name": "Vasily Romanov",
			"status":    20,
		},
	}
w.WriteHeader(http.StatusOK)
jsonResp, _ := json.Marshal(result)
w.Write(jsonResp)
}
`))
)

map[MyApi:[map[Profile:{URL:/user/profile Auth:false Method:}] map[Create:{URL:/user/create Auth:true Method:POST}]]]