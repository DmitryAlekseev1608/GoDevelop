package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"regexp"
	"text/template"
)

type parsingResult map[string][]method

type method map[string]methodURLPath

type methodURLPath struct {
	URL string `json:"url"`
	Auth bool `json:"auth"`
	Method string `json:"method"`
}

var (
	serveHTTPHeaderTpl = template.Must(template.New("serveHTTPHeaderTpl").Parse(`
func (h *{{.StructName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
switch r.URL.Path {`))
	serveHTTPBodyTpl = template.Must(template.New("serveHTTPBodyTpl").Parse(`
	case "{{.URLPath}}":
		h.handler{{.MethodName}}(w, r)`))
	handlerTpl = template.Must(template.New("handlerTpl").Parse(`
func (h *{{.StructName}}) handler{{.MethodName}} (w http.ResponseWriter, r *http.Request) {
	res, err := h.{{.MethodName}}(ctx, params)
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
	)`)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `type resp map[string]interface{}`)

	parsingResult := parsingResult{}
	for _, f := range node.Decls {
		method := method{}
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
				if needCodegen {
					re := regexp.MustCompile(`\{.*\}`)
					var parsingMethod methodURLPath
					err := json.Unmarshal([]byte(re.FindString(comment.Text)), &parsingMethod)
					if err != nil {
						log.Fatal(err)
					}
					method[f.(*ast.FuncDecl).Name.Name] = parsingMethod
					break
					// fmt.Printf("type: %T data: %+v\n", parsingMethod, parsingMethod)
				}
			}
			if !needCodegen {
				fmt.Printf("SKIP method %#v doesnt have apigen mark\n", g.Name)
				continue
			}
		if starExpr, ok := g.Recv.List[0].Type.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				parsingResult[ident.Name] = append(parsingResult[ident.Name], method)
			}
		}
		fmt.Printf("type: %T data: %+v\n", parsingResult, parsingResult)
	}

	for str, paramStr := range parsingResult {
		dataPostTpl := map[string] interface{} {
			"StructName":   str,
		}
		serveHTTPHeaderTpl.Execute(out, dataPostTpl)
		for _, paramMet := range paramStr {
			for function, param := range paramMet {
				dataPostTpl := map[string] interface{} {
					"URLPath":   param.URL,
					"MethodName": function,
				}
				serveHTTPBodyTpl.Execute(out, dataPostTpl)
			}
		}
		fmt.Fprintln(out, `
			}`)
		fmt.Fprintln(out, `}`)
		for _, paramMet := range paramStr {
			for function, _ := range paramMet {
				dataPostTpl := map[string] interface{} {
					"StructName":   str,
					"MethodName": function,
				}
				handlerTpl.Execute(out, dataPostTpl)
			}
		}
	}
}
