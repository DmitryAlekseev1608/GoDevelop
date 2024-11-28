package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

type tpl struct {
	FieldName string
}


var (
	intTpl = template.Must(template.New("intTpl").Parse(`
// {{.FieldName}}
func (h *{{.FieldName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

// func (h *{{.FieldName}}) wrapperDoSomeJob() {
// 	res, err := h.DoSomeJob(ctx, params)
}
`))
)


func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out, 
	`
	import (
		"net/http"
		"encoding/json"

	)`)

	fmt.Fprintln(out, `type resp map[string]interface{}`)

	methodName := make(map[string]struct{})
	for _, f := range node.Decls {
		g, ok := f.(*ast.FuncDecl)
		if !ok {
			fmt.Printf("SKIP %#T is not *ast.FuncDecl\n", f)
			continue
		}
		if g.Recv == nil {
			fmt.Printf("SKIP functions %#v\n", g.Name)
			continue
		}
		if g.Doc == nil {
			fmt.Printf("SKIP method %#v doesnt have comments\n", g.Name)
			continue
		}
		needCodegen := false
		for _, comment := range g.Doc.List {
				needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// apigen:api")
			}
			if !needCodegen {
				fmt.Printf("SKIP method %#v doesnt have cgen mark\n", g.Name)
				continue
			}
		if starExpr, ok := g.Recv.List[0].Type.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				methodName[ident.Name] = struct{}{}
			}
		}
	}

	// fmt.Printf("type: %T data: %+v\n", methodName[0], methodName[0])

	for fieldName, _ := range methodName {
		intTpl.Execute(out, tpl{fieldName})
	}
}
