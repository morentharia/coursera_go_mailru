package main

// код писать тут
// go build gen/* && ./codegen.exe pack/unpack.go  pack/marshaller.go
// go run pack/*

// ➜   hw5_codegen git:(master) ✗ ag -l | entr -c bash -c "go build handlers_gen/* && ./codegen api.go api_handlers.go"
// ➜   hw5_codegen git:(master) ✗ ls api_handlers.go | entr -c bash -c "cat api_handlers.go && go test -v -run TestMyApi"
import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"
)

var output string

type StructField struct {
	Name     string
	TypeName string
	Tag      string
}

type typeParsed struct {
	Name   string
	Fields []StructField
}

type FuncComment struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method,omitempty"`
}

type funcParsed struct {
	Comment      FuncComment
	RecvName     string
	RecvType     string
	FunctionName string
	ParamsType   string
	ResultType   string
}

func parseKeyVal(pairStr string) (key, value string) {
	res := strings.Split(pairStr, "=")
	if len(res) == 2 {
		key, value = res[0], res[1]
	} else if len(res) == 1 {
		key, value = res[0], ""
	}
	return
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	funcParsedList := make([]funcParsed, 0)
	typeName2Parsed := make(map[string]*typeParsed)

	for _, f := range node.Decls {
		switch f.(type) {
		case *ast.GenDecl:
			g, _ := f.(*ast.GenDecl)

			for _, spec := range g.Specs {
				currType, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				currStruct, ok := currType.Type.(*ast.StructType)
				if !ok {
					continue
				}

				typeName2Parsed[currType.Name.Name] = &typeParsed{
					Name:   currType.Name.Name,
					Fields: make([]StructField, 0),
				}
				for _, field := range currStruct.Fields.List {
					if field.Tag != nil {
						tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])

						if tag.Get("apivalidator") == "-" || tag.Get("apivalidator") == "" {
							continue
						}
						fieldName := field.Names[0].Name
						fieldType := field.Type.(*ast.Ident).Name

						typeName2Parsed[currType.Name.Name].Fields = append(
							typeName2Parsed[currType.Name.Name].Fields,
							StructField{
								Name:     fieldName,
								TypeName: fieldType,
								Tag:      field.Tag.Value,
							},
						)

					}
				}

			}
		case *ast.FuncDecl:
			function, _ := f.(*ast.FuncDecl)

			comment := function.Doc.Text()
			if !strings.HasPrefix(comment, "apigen:api") {
				continue
			}

			fg := funcParsed{}

			comment = strings.TrimPrefix(comment, "apigen:api")
			err := json.Unmarshal([]byte(comment), &fg.Comment)
			if err != nil {
				fmt.Errorf("Cant decode json")
				continue
			}

			fg.FunctionName = function.Name.Name
			fg.RecvName = function.Recv.List[0].Names[0].Name
			fg.RecvType = function.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
			for _, field := range function.Type.Params.List {
				ident, ok := field.Type.(*ast.Ident)
				if ok {
					fg.ParamsType = ident.Name
					break
				}
			}
			for _, field := range function.Type.Results.List {
				zzz, ok := field.Type.(*ast.StarExpr)
				if ok {
					fg.ResultType = zzz.X.(*ast.Ident).Name
					break
				}
			}
			funcParsedList = append(funcParsedList, fg)
		}
	}

	libsTpl := template.Must(template.New("headerTpl").Parse(`
package main
import (
	"net/http"
	"strconv"
	"encoding/json"
	"strings"
	"fmt"
)
	`))

	headerTpl := template.Must(template.New("headerTpl").Parse(`
func (h *{{.RecvType}} ) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	{{- range .Routes}}
	case "{{.Comment.Url}}":
		h.{{.FunctionName}}Handler(w, r)
	{{- end}}
	default:
		// 404
		w.WriteHeader(http.StatusNotFound)
		mk := make(map[string]interface{})
		mk["error"] = "unknown method"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
	}
}
	`))

	funcTplHeader := template.Must(template.New("funcTplHeader").Parse(`
// URL: {{.Comment.Url}}
func (h *{{.RecvType}}) {{.FunctionName}}Handler(w http.ResponseWriter, r *http.Request) {
	var err error
	tmp := ""
	fmt.Printf("tmp = %+v\n", tmp)
	params:=new({{.ParamsType}})
`))
	onlyPOSTTplInt := template.Must(template.New("onlyPOSTTplInt").Parse(`
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		mk := make(map[string]interface{})
		mk["error"] = "bad method"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	`))
	onlyAuthTplInt := template.Must(template.New("onlyAuthTplInt").Parse(`
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		mp := make(map[string]interface{})
		mp["error"] = "unauthorized"
		js, _ := json.Marshal(mp)
		w.Write(js)
		return
	}
	`))
	paramTplInt := template.Must(template.New("paramTplInt").Parse(`
	if r.Method == "POST" {
		tmp = r.FormValue(strings.ToLower("{{.ParamName}}"))
	} else {
		tmp = r.URL.Query().Get(strings.ToLower("{{.ParamName}}"))
	}
	params.{{.Name}}, err = strconv.Atoi(tmp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("{{.Name}}") + " must be int"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	`))

	paramTplStr := template.Must(template.New("paramTplStr").Parse(`
	if r.Method == "POST" {
		params.{{.Name}} = r.FormValue(strings.ToLower("{{.ParamName}}"))
	} else {
		params.{{.Name}} = r.URL.Query().Get(strings.ToLower("{{.ParamName}}"))
	}
	`))

	fieldRequiredTpl := template.Must(template.New("fieldTpl").Parse(`
	if params.{{.Name}} == "" {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("{{.Name}}") + " must me not empty"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	`))

	fieldMinString := template.Must(template.New("fieldTpl").Parse(`
	if !(len(params.{{.Name}}) >= {{.Min}}) {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("{{.Name}}") + " len must be >= {{.Min}}"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	`))

	fieldMinInt := template.Must(template.New("fieldTpl").Parse(`
	if !(params.{{.Name}} >= {{.Min}}) {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("{{.Name}}") + " must be >= {{.Min}}"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	`))

	fieldMaxString := template.Must(template.New("fieldTpl").Parse(`
	if !(len(params.{{.Name}}) < {{.Max}}) {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("{{.Name}}") + " len must be <= {{.Max}}"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	`))

	fieldMaxInt := template.Must(template.New("fieldTpl").Parse(`
	if !(params.{{.Name}} < {{.Max}}) {
		w.WriteHeader(http.StatusBadRequest)
		mk := make(map[string]interface{})
		mk["error"] = strings.ToLower("{{.Name}}") + " must be <= {{.Max}}"
		resp, _ := json.Marshal(mk)
		w.Write(resp)
		return
	}
	`))

	var fns = template.FuncMap{
		"plus1": func(x int) int {
			return x + 1
		},
	}
	fieldEnumTpl := template.Must(template.New("fieldEnumTpl").Funcs(fns).Parse(`
	{{$n := len .Cases}}
	switch params.{{.Name}} {
		{{- range .Cases}}
		case "{{.}}":
	    {{- end}}
		// params.{{.Name}} = {{.Name}}
	default:
		if params.{{.Name}} != "" {
			w.WriteHeader(http.StatusBadRequest)
			mk := make(map[string]interface{})
			mk["error"] = strings.ToLower("{{.Name}}") + " must be one of [{{- range $i, $e := .Cases}}{{.}}{{if eq (plus1 $i) $n}}]{{ else }}, {{end}}{{- end}}"
			resp, _ := json.Marshal(mk)
			w.Write(resp)
			return
		}
		params.{{.Name}} = "{{.Default}}"
	}`))

	callFuncTpl := template.Must(template.New("callFuncTpl").Parse(`
	ctx := r.Context()

	res, err := h.{{.FunctionName}}(ctx, *params)
	if err != nil {
		e, ok := err.(ApiError)
		if ok {
			w.WriteHeader(e.HTTPStatus)
			mk := make(map[string]interface{})
			mk["error"] = err.Error()
			resp, _ := json.Marshal(mk)
			w.Write(resp)
			return
		} else {
			if err != nil && err.Error() == "bad user" {
				w.WriteHeader(http.StatusInternalServerError)
				mk := make(map[string]interface{})
				mk["error"] = err.Error()
				resp, _ := json.Marshal(mk)
				w.Write(resp)
				return
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	mk := make(map[string]interface{})
	mk["error"] =	 ""
	mk["response"] = res
	resp, _ := json.Marshal(mk)
	w.Write(resp)
	`))

	funcTplFooter := template.Must(template.New("funcTplFooter").Parse(`
}
`))

	re := regexp.MustCompile(`apivalidator:"(.*)"`)
	libsTpl.Execute(out, struct{}{})

	groupByRecvType := make(map[string][]funcParsed)
	for _, funcParsed := range funcParsedList {
		groupByRecvType[funcParsed.RecvType] = append(groupByRecvType[funcParsed.RecvType], funcParsed)
	}
	for RecvType, Routes := range groupByRecvType {
		headerTpl.Execute(
			out,
			struct {
				RecvType string
				Routes   []funcParsed
			}{RecvType, Routes},
		)
	}
	for _, funcParsed := range funcParsedList {
		funcTplHeader.Execute(out, funcParsed)
		params := typeName2Parsed[funcParsed.ParamsType]

		if funcParsed.Comment.Method == "POST" {
			onlyPOSTTplInt.Execute(out, struct{}{})
		}
		if funcParsed.Comment.Auth {
			onlyAuthTplInt.Execute(out, struct{}{})
		}
		for _, field := range params.Fields {
			var paramName string

			tagStr := re.FindStringSubmatch(field.Tag)[1]
			for _, val := range strings.Split(tagStr, ",") {
				key, value := parseKeyVal(val)
				if key == "paramname" {
					paramName = value
				}
			}
			if paramName == "" {
				paramName = field.Name
			}

			if field.TypeName == "int" {
				paramTplInt.Execute(
					out,
					struct {
						Name      string
						ParamName string
					}{
						Name:      field.Name,
						ParamName: paramName,
					},
				)
			} else {
				paramTplStr.Execute(out,
					struct {
						Name      string
						ParamName string
					}{
						Name:      field.Name,
						ParamName: paramName,
					},
				)
			}

		}
		for _, field := range params.Fields {
			var defaultVal string

			tagStr := re.FindStringSubmatch(field.Tag)[1]
			for _, val := range strings.Split(tagStr, ",") {
				key, value := parseKeyVal(val)
				if key == "default" {
					defaultVal = value
				}
			}

			for _, val := range strings.Split(tagStr, ",") {
				key, value := parseKeyVal(val)

				switch key {
				case "required":
					fieldRequiredTpl.Execute(
						out,
						struct {
							Name string
						}{
							Name: field.Name,
						},
					)
				case "enum":
					fieldEnumTpl.Execute(out, struct {
						Name    string
						Cases   []string
						Default string
					}{
						Name:    field.Name,
						Cases:   strings.Split(value, "|"),
						Default: defaultVal,
					},
					)
				case "min":
					if field.TypeName == "string" {
						fieldMinString.Execute(
							out,
							struct {
								Name string
								Min  string
							}{field.Name, value},
						)
					} else if field.TypeName == "int" {
						fieldMinInt.Execute(
							out,
							struct {
								Name string
								Min  string
							}{field.Name, value},
						)
					}
				case "max":
					if field.TypeName == "string" {
						fieldMaxString.Execute(
							out,
							struct {
								Name string
								Max  string
							}{field.Name, value},
						)
					} else if field.TypeName == "int" {
						fieldMaxInt.Execute(
							out,
							struct {
								Name string
								Max  string
							}{field.Name, value},
						)
					}
				}
			}
		}
		callFuncTpl.Execute(out, funcParsed)
		funcTplFooter.Execute(out, funcParsed)
	}
}
