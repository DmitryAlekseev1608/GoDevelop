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


var (
	intTpl = template.Must(template.New("intTpl").Parse(`
	// {{.FieldName}}
	type {{.FieldName}} struct{}
	func (h *{{.FieldName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user/profile":
			fmt.Println("-----------")
		}
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
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import "encoding/binary"`)
	fmt.Fprintln(out, `import "bytes"`)
	fmt.Fprintln(out) // empty line

	methodname := make([]ast.Expr, 0)
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
		methodname = append(methodname, g.Recv.List[0].Type)
	}
	fmt.Println(methodname)
}
