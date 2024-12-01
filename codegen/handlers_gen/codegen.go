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
	Attributes string `json:"attribute"`
}

type validParam map[string]valid 

type valid struct {
	Required bool
	Paramname string
	Enum []string
	Default interface{}
	Min interface{}
	Max interface{}
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
	ctx := r.Context()
	params := {{.AtrebutesName}}{}
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
	res, err := h.{{.MethodName}}(ctx, params)
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
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/schema"
	"fmt"
	)`)
	fmt.Fprintln(out)
	fmt.Fprintln(out, `type resp map[string]interface{}`)

	parsingResult := parsingResult{}
	needStruct := make([]string, 0)
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
					if secondAttributeName, ok := g.Type.Params.List[1].Type.(*ast.Ident); ok {
						parsingMethod.Attributes = secondAttributeName.Name
						needStruct = append(needStruct, secondAttributeName.Name)
					}
					method[f.(*ast.FuncDecl).Name.Name] = parsingMethod
					break
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
		// fmt.Printf("type: %T data: %+v\n", parsingResult, parsingResult)
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
			for function, paramFunc := range paramMet {
				dataPostTpl := map[string] interface{} {
					"StructName":   str,
					"MethodName": function,
					"AtrebutesName": paramFunc.Attributes,
				}
				handlerTpl.Execute(out, dataPostTpl)
			}
		}
	}


	validParam := make(validParam)
	for _, f := range node.Decls {
		g, ok := f.(*ast.GenDecl)
		if !ok {
			fmt.Printf("SKIP %#T is not *ast.GenDecl\n", f)
			continue
		}
		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				fmt.Printf("SKIP %#T is not ast.TypeSpec\n", spec)
				continue
			}
			if contains(needStruct, currType.Name.Name) {
				if currStruct, ok := currType.Type.(*ast.StructType); ok {
					for _, tag := range currStruct.Fields.List {
						valid := valid{
							Required: strings.Contains(tag.Tag.Value, "required"),
							Paramname: takeParamName(tag.Tag.Value),
							Enum: takeEnum(tag.Tag.Value),
							Default: takeDefault(tag.Tag.Value),
							Min: takeMin(tag.Tag.Value),
							Max: takeMax(tag.Tag.Value),
						}
						validParam[tag.Names[0].Name] = valid
						// fmt.Printf("type: %T data: %+v\n", currStruct.Fields.List[0].Names[0].Name, currStruct.Fields.List[0].Names[0].Name)
						// fmt.Printf("type: %T data: %+v\n", currStruct.Fields.List[0].Tag.Kind, currStruct.Fields.List[0].Tag.Kind)
						// fmt.Printf("type: %T data: %+v\n", currStruct.Fields.List[0].Tag.Value, currStruct.Fields.List[0].Tag.Value)
					}
				}
			}
		}
	}
	fmt.Printf("type: %T data: %+v\n", validParam, validParam)
}

func contains(slice []string, value string) bool {
    for _, v := range slice {
        if v == value {
            return true
        }
    }
    return false
}

func takeParamName (tag string) string {
	pattern := `paramname=([a-zA-Z_]+)`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(tag)
	if len(match) > 0 {
		return match[1]
	} else {
		return ""
	}
}

func takeEnum (tag string) []string {
	pattern := `enum=([a-zA-Z_]+)`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(tag)
	if len(match) > 0 {
		return strings.Split(match[1], "|")
	} else {
		return nil
	}
}

func takeDefault (tag string) string {
	pattern := `default=([a-zA-Z_]+)`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(tag)
	if len(match) > 0 {
		return match[1]
	} else {
		return ""
	}
}

func takeMin (tag string) string {
	pattern := `min=([0-9]+)`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(tag)
	if len(match) > 0 {
		return match[1]
	} else {
		return ""
	}
}

func takeMax (tag string) string {
	pattern := `max=([0-9]+)`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(tag)
	if len(match) > 0 {
		return match[1]
	} else {
		return ""
	}
}