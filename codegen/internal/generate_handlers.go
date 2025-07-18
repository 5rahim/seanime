package codegen

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type (
	RouteHandler struct {
		Name        string           `json:"name"`
		TrimmedName string           `json:"trimmedName"`
		Comments    []string         `json:"comments"`
		Filepath    string           `json:"filepath"`
		Filename    string           `json:"filename"`
		Api         *RouteHandlerApi `json:"api"`
	}

	RouteHandlerApi struct {
		Summary              string               `json:"summary"`
		Descriptions         []string             `json:"descriptions"`
		Endpoint             string               `json:"endpoint"`
		Methods              []string             `json:"methods"`
		Params               []*RouteHandlerParam `json:"params"`
		BodyFields           []*RouteHandlerParam `json:"bodyFields"`
		Returns              string               `json:"returns"`
		ReturnGoType         string               `json:"returnGoType"`
		ReturnTypescriptType string               `json:"returnTypescriptType"`
	}

	RouteHandlerParam struct {
		Name             string   `json:"name"`
		JsonName         string   `json:"jsonName"`
		GoType           string   `json:"goType"`                     // e.g., []models.User
		InlineStructType string   `json:"inlineStructType,omitempty"` // e.g., struct{Test string `json:"test"`}
		UsedStructType   string   `json:"usedStructType"`             // e.g., models.User
		TypescriptType   string   `json:"typescriptType"`             // e.g., Array<User>
		Required         bool     `json:"required"`
		Descriptions     []string `json:"descriptions"`
	}
)

func GenerateHandlers(dir string, outDir string) {

	handlers := make([]*RouteHandler, 0)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".go") || strings.HasPrefix(info.Name(), "_") {
			return nil
		}

		// Parse the file
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, decl := range file.Decls {
			// Check if the declaration is a function
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			// Check if the function has comments
			if fn.Doc == nil {
				continue
			}

			// Get the comments
			comments := strings.Split(fn.Doc.Text(), "\n")
			if len(comments) == 0 {
				continue
			}

			// Get the function name
			name := fn.Name.Name
			trimmedName := strings.TrimPrefix(name, "Handle")

			// Get the filename
			filep := strings.ReplaceAll(strings.ReplaceAll(path, "\\", "/"), "../", "")
			filename := filepath.Base(path)

			// Get the endpoint
			endpoint := ""
			var methods []string
			params := make([]*RouteHandlerParam, 0)
			summary := ""
			descriptions := make([]string, 0)
			returns := "bool"

			for _, comment := range comments {
				cmt := strings.TrimSpace(strings.TrimPrefix(comment, "//"))
				if strings.HasPrefix(cmt, "@summary") {
					summary = strings.TrimSpace(strings.TrimPrefix(cmt, "@summary"))
				}

				if strings.HasPrefix(cmt, "@desc") {
					descriptions = append(descriptions, strings.TrimSpace(strings.TrimPrefix(cmt, "@desc")))
				}

				if strings.HasPrefix(cmt, "@route") {
					endpointParts := strings.Split(strings.TrimSpace(strings.TrimPrefix(cmt, "@route")), " ")
					if len(endpointParts) == 2 {
						endpoint = endpointParts[0]
						methods = strings.Split(endpointParts[1][1:len(endpointParts[1])-1], ",")
					}
				}

				if strings.HasPrefix(cmt, "@param") {
					paramParts := strings.Split(strings.TrimSpace(strings.TrimPrefix(cmt, "@param")), " - ")
					if len(paramParts) == 4 {
						required := paramParts[2] == "true"
						params = append(params, &RouteHandlerParam{
							Name:           paramParts[0],
							JsonName:       paramParts[0],
							GoType:         paramParts[1],
							TypescriptType: goTypeToTypescriptType(paramParts[1]),
							Required:       required,
							Descriptions:   []string{strings.ReplaceAll(paramParts[3], "\"", "")},
						})
					}
				}

				if strings.HasPrefix(cmt, "@returns") {
					returns = strings.TrimSpace(strings.TrimPrefix(cmt, "@returns"))
				}
			}

			bodyFields := make([]*RouteHandlerParam, 0)
			// To get the request body fields, we need to look at the function body for a struct called "body"

			// Get the function body
			body := fn.Body
			if body != nil {
				for _, stmt := range body.List {
					// Check if the statement is a declaration
					declStmt, ok := stmt.(*ast.DeclStmt)
					if !ok {
						continue
					}
					// Check if the declaration is a gen decl
					genDecl, ok := declStmt.Decl.(*ast.GenDecl)
					if !ok {
						continue
					}
					// Check if the declaration is a type
					if genDecl.Tok != token.TYPE {
						continue
					}
					// Check if the type is a struct
					if len(genDecl.Specs) != 1 {
						continue
					}
					typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)
					if !ok {
						continue
					}
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					// Check if the struct is called "body"
					if typeSpec.Name.Name != "body" {
						continue
					}

					// Get the fields
					for _, field := range structType.Fields.List {
						// Get the field name
						fieldName := field.Names[0].Name

						// Get the field type
						fieldType := field.Type

						jsonName := fieldName
						// Get the field tag
						required := !jsonFieldOmitEmpty(field)
						jsonField := jsonFieldName(field)
						if jsonField != "" {
							jsonName = jsonField
						}

						// Get field comments
						fieldComments := make([]string, 0)
						cmtsTxt := field.Doc.Text()
						if cmtsTxt != "" {
							fieldComments = strings.Split(cmtsTxt, "\n")
						}
						for _, cmt := range fieldComments {
							cmt = strings.TrimSpace(strings.TrimPrefix(cmt, "//"))
							if cmt != "" {
								fieldComments = append(fieldComments, cmt)
							}
						}

						switch fieldType.(type) {
						case *ast.StarExpr:
							required = false
						}

						goType := fieldTypeString(fieldType)
						goTypeUnformatted := fieldTypeUnformattedString(fieldType)
						packageName := "handlers"
						if strings.Contains(goTypeUnformatted, ".") {
							parts := strings.Split(goTypeUnformatted, ".")
							packageName = parts[0]
						}

						tsType := fieldTypeToTypescriptType(fieldType, packageName)

						usedStructType := goTypeUnformatted
						switch goTypeUnformatted {
						case "string", "int", "int64", "float64", "float32", "bool", "nil", "uint", "uint64", "uint32", "uint16", "uint8", "byte", "rune", "[]byte", "interface{}", "error":
							usedStructType = ""
						}

						// Add the request body field
						bodyFields = append(bodyFields, &RouteHandlerParam{
							Name:           fieldName,
							JsonName:       jsonName,
							GoType:         goType,
							UsedStructType: usedStructType,
							TypescriptType: tsType,
							Required:       required,
							Descriptions:   fieldComments,
						})

						// Check if it's an inline struct and capture its definition
						if structType, ok := fieldType.(*ast.StructType); ok {
							bodyFields[len(bodyFields)-1].InlineStructType = formatInlineStruct(structType)
						} else {
							// Check if it's a slice of inline structs
							if arrayType, ok := fieldType.(*ast.ArrayType); ok {
								if structType, ok := arrayType.Elt.(*ast.StructType); ok {
									bodyFields[len(bodyFields)-1].InlineStructType = "[]" + formatInlineStruct(structType)
								}
							}
							// Check if it's a map with inline struct values
							if mapType, ok := fieldType.(*ast.MapType); ok {
								if structType, ok := mapType.Value.(*ast.StructType); ok {
									bodyFields[len(bodyFields)-1].InlineStructType = "map[" + fieldTypeString(mapType.Key) + "]" + formatInlineStruct(structType)
								}
							}
						}
					}
				}
			}

			// Add the route handler
			routeHandler := &RouteHandler{
				Name:        name,
				TrimmedName: trimmedName,
				Comments:    comments,
				Filepath:    filep,
				Filename:    filename,
				Api: &RouteHandlerApi{
					Summary:              summary,
					Descriptions:         descriptions,
					Endpoint:             endpoint,
					Methods:              methods,
					Params:               params,
					BodyFields:           bodyFields,
					Returns:              returns,
					ReturnGoType:         getUnformattedGoType(returns),
					ReturnTypescriptType: stringGoTypeToTypescriptType(returns),
				},
			}

			handlers = append(handlers, routeHandler)

		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	// Write structs to file
	_ = os.MkdirAll(outDir, os.ModePerm)
	file, err := os.Create(outDir + "/handlers.json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(handlers); err != nil {
		fmt.Println("Error:", err)
		return
	}

	return
}
