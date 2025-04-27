package codegen

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"
)

type GoStruct struct {
	Filepath            string           `json:"filepath"`
	Filename            string           `json:"filename"`
	Name                string           `json:"name"`
	FormattedName       string           `json:"formattedName"` // name with package prefix e.g. models.User => Models_User
	Package             string           `json:"package"`
	Fields              []*GoStructField `json:"fields"`
	AliasOf             *GoAlias         `json:"aliasOf,omitempty"`
	Comments            []string         `json:"comments"`
	EmbeddedStructTypes []string         `json:"embeddedStructNames,omitempty"`
}

type GoAlias struct {
	GoType             string   `json:"goType"`
	TypescriptType     string   `json:"typescriptType"`
	UsedTypescriptType string   `json:"usedTypescriptType,omitempty"`
	DeclaredValues     []string `json:"declaredValues"`
	UsedStructType     string   `json:"usedStructName,omitempty"`
}

type GoStructField struct {
	Name     string `json:"name"`
	JsonName string `json:"jsonName"`
	// e.g. map[string]models.User
	GoType string `json:"goType"`
	// e.g. User
	TypescriptType string `json:"typescriptType"`
	// e.g. TypescriptType = Array<Models_User> => UsedTypescriptType = Models_User
	UsedTypescriptType string `json:"usedTypescriptType,omitempty"`
	// e.g. GoType = map[string]models.User => TypescriptType = User => UsedStructType = models.User
	UsedStructType string `json:"usedStructName,omitempty"`
	// If no 'omitempty' and not a pointer
	Required bool     `json:"required"`
	Public   bool     `json:"public"`
	Comments []string `json:"comments"`
}

var typePrefixesByPackage = map[string]string{
	"anilist":                "AL_",
	"auto_downloader":        "AutoDownloader_",
	"autodownloader":         "AutoDownloader_",
	"entities":               "",
	"db":                     "DB_",
	"db_bridge":              "DB_",
	"models":                 "Models_",
	"playbackmanager":        "PlaybackManager_",
	"torrent_client":         "TorrentClient_",
	"events":                 "Events_",
	"torrent":                "Torrent_",
	"manga":                  "Manga_",
	"autoscanner":            "AutoScanner_",
	"listsync":               "ListSync_",
	"util":                   "Util_",
	"scanner":                "Scanner_",
	"offline":                "Offline_",
	"discordrpc":             "DiscordRPC_",
	"discordrpc_presence":    "DiscordRPC_",
	"anizip":                 "Anizip_",
	"onlinestream":           "Onlinestream_",
	"onlinestream_providers": "Onlinestream_",
	"onlinestream_sources":   "Onlinestream_",
	"manga_providers":        "Manga_",
	"chapter_downloader":     "ChapterDownloader_",
	"manga_downloader":       "MangaDownloader_",
	"docs":                   "INTERNAL_",
	"tvdb":                   "TVDB_",
	"metadata":               "Metadata_",
	"mappings":               "Mappings_",
	"mal":                    "MAL_",
	"handlers":               "",
	"animetosho":             "AnimeTosho_",
	"updater":                "Updater_",
	"anime":                  "Anime_",
	"anime_types":            "Anime_",
	"summary":                "Summary_",
	"filesystem":             "Filesystem_",
	"filecache":              "Filecache_",
	"core":                   "INTERNAL_",
	"comparison":             "Comparison_",
	"mediastream":            "Mediastream_",
	"torrentstream":          "Torrentstream_",
	"extension":              "Extension_",
	"extension_repo":         "ExtensionRepo_",
	//"vendor_hibike_manga":        "HibikeManga_",
	//"vendor_hibike_onlinestream": "HibikeOnlinestream_",
	//"vendor_hibike_torrent":      "HibikeTorrent_",
	//"vendor_hibike_mediaplayer":  "HibikeMediaPlayer_",
	//"vendor_hibike_extension":    "HibikeExtension_",
	"hibikemanga":        "HibikeManga_",
	"hibikeonlinestream": "HibikeOnlinestream_",
	"hibiketorrent":      "HibikeTorrent_",
	"hibikemediaplayer":  "HibikeMediaPlayer_",
	"hibikeextension":    "HibikeExtension_",
	"continuity":         "Continuity_",
	"sync":               "Sync_",
	"debrid":             "Debrid_",
	"debrid_client":      "DebridClient_",
	"report":             "Report_",
	"habari":             "Habari_",
	"vendor_habari":      "Habari_",
	"discordrpc_client":  "DiscordRPC_",
}

func getTypePrefix(packageName string) string {
	if prefix, ok := typePrefixesByPackage[packageName]; ok {
		return prefix
	}
	return ""
}

func ExtractStructs(dir string, outDir string) {

	structs := make([]*GoStruct, 0)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			res, err := getGoStructsFromFile(path, info)
			if err != nil {
				return err
			}
			structs = append(structs, res...)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write structs to file
	_ = os.MkdirAll(outDir, os.ModePerm)
	file, err := os.Create(outDir + "/public_structs.json")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(structs); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Public structs extracted and saved to public_structs.json")
}

func getGoStructsFromFile(path string, info os.FileInfo) (structs []*GoStruct, err error) {

	// Parse the Go file
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	packageName := file.Name.Name

	// Extract public structs
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		//
		// Go through each type declaration
		//
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if !typeSpec.Name.IsExported() {
				continue
			}

			//
			// The type declaration is an alias
			// e.g. alias.Name: string, typeSpec.Name.Name: MediaListStatus
			//
			alias, ok := typeSpec.Type.(*ast.Ident)
			if ok {

				if alias.Name == typeSpec.Name.Name {
					continue
				}
				goStruct := goStructFromAlias(path, info, genDecl, typeSpec, packageName, alias, file)
				structs = append(structs, goStruct)
				continue
			}

			//
			// The type declaration is a struct
			//
			structType, ok := typeSpec.Type.(*ast.StructType)
			if ok {

				subStructs := make([]*GoStruct, 0)
				for _, field := range structType.Fields.List {
					if field.Names != nil && len(field.Names) > 0 {

						subStructType, ok := field.Type.(*ast.StructType)
						if ok {
							name := fmt.Sprintf("%s_%s", typeSpec.Name.Name, field.Names[0].Name)
							subStruct := goStructFromStruct(path, info, genDecl, name, packageName, subStructType)
							subStructs = append(subStructs, subStruct)
							continue
						}

					}
				}

				goStruct := goStructFromStruct(path, info, genDecl, typeSpec.Name.Name, packageName, structType)

				// Replace struct fields with sub structs
				for _, field := range goStruct.Fields {
					if field.GoType == "__STRUCT__" {
						for _, subStruct := range subStructs {
							if subStruct.Name == fmt.Sprintf("%s_%s", typeSpec.Name.Name, field.Name) {
								field.GoType = subStruct.FormattedName
								field.TypescriptType = subStruct.FormattedName
								field.UsedStructType = fmt.Sprintf("%s.%s", subStruct.Package, subStruct.Name)
								break
							}
						}
					}
				}

				structs = append(structs, goStruct)
				structs = append(structs, subStructs...)
				continue
			}

			mapType, ok := typeSpec.Type.(*ast.MapType)
			if ok {
				goStruct := &GoStruct{
					Filepath:      path,
					Filename:      info.Name(),
					Name:          typeSpec.Name.Name,
					FormattedName: getTypePrefix(packageName) + typeSpec.Name.Name,
					Package:       packageName,
					Fields:        make([]*GoStructField, 0),
				}

				usedStructType, usedStructPkgName := getUsedStructType(mapType, packageName)

				goStruct.AliasOf = &GoAlias{
					GoType:         fieldTypeString(mapType),
					TypescriptType: fieldTypeToTypescriptType(mapType, usedStructPkgName),
					UsedStructType: usedStructType,
				}

				structs = append(structs, goStruct)
				continue
			}

			sliceType, ok := typeSpec.Type.(*ast.ArrayType)
			if ok {
				goStruct := &GoStruct{
					Filepath:      path,
					Filename:      info.Name(),
					Name:          typeSpec.Name.Name,
					FormattedName: getTypePrefix(packageName) + typeSpec.Name.Name,
					Package:       packageName,
					Fields:        make([]*GoStructField, 0),
				}

				usedStructType, usedStructPkgName := getUsedStructType(sliceType, packageName)

				goStruct.AliasOf = &GoAlias{
					GoType:         fieldTypeString(sliceType),
					TypescriptType: fieldTypeToTypescriptType(sliceType, usedStructPkgName),
					UsedStructType: usedStructType,
				}

				structs = append(structs, goStruct)
				continue
			}

		}
	}
	return structs, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Example:
//
//	type User struct {
//		ID   int    `json:"id"`
//		Name string `json:"name"`
//	}
func goStructFromStruct(path string, info os.FileInfo, genDecl *ast.GenDecl, name string, packageName string, structType *ast.StructType) *GoStruct {
	// Get comments
	comments := make([]string, 0)
	if genDecl.Doc != nil && genDecl.Doc.List != nil && len(genDecl.Doc.List) > 0 {
		for _, comment := range genDecl.Doc.List {
			comments = append(comments, strings.TrimPrefix(comment.Text, "//"))
		}
	}

	goStruct := &GoStruct{
		Filepath:            filepath.ToSlash(path),
		Filename:            info.Name(),
		Name:                name,
		FormattedName:       getTypePrefix(packageName) + name,
		Package:             packageName,
		Fields:              make([]*GoStructField, 0),
		EmbeddedStructTypes: make([]string, 0),
		Comments:            comments,
	}

	// Get fields
	for _, field := range structType.Fields.List {
		if field.Names == nil || len(field.Names) == 0 {
			if len(field.Names) == 0 {
				switch field.Type.(type) {
				case *ast.Ident, *ast.StarExpr, *ast.SelectorExpr:
					usedStructType, _ := getUsedStructType(field.Type, packageName)
					goStruct.EmbeddedStructTypes = append(goStruct.EmbeddedStructTypes, usedStructType)
				}
			}
			continue
		}
		// Get fields comments
		comments := make([]string, 0)
		if field.Comment != nil && field.Comment.List != nil && len(field.Comment.List) > 0 {
			for _, comment := range field.Comment.List {
				comments = append(comments, strings.TrimPrefix(comment.Text, "//"))
			}
		}

		required := true
		if field.Tag != nil {
			tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
			jsonTag := tag.Get("json")
			if jsonTag != "" {
				jsonParts := strings.Split(jsonTag, ",")
				if len(jsonParts) > 1 && jsonParts[1] == "omitempty" {
					required = false
				}
			}
		}
		switch field.Type.(type) {
		case *ast.StarExpr, *ast.ArrayType, *ast.MapType, *ast.SelectorExpr:
			required = false
		}
		fieldName := field.Names[0].Name

		usedStructType, usedStructPkgName := getUsedStructType(field.Type, packageName)

		tsType := fieldTypeToTypescriptType(field.Type, usedStructPkgName)

		goStructField := &GoStructField{
			Name:               fieldName,
			JsonName:           jsonFieldName(field),
			GoType:             fieldTypeString(field.Type),
			TypescriptType:     tsType,
			UsedTypescriptType: fieldTypeToUsedTypescriptType(tsType),
			Required:           required,
			Public:             field.Names[0].IsExported(),
			UsedStructType:     usedStructType,
			Comments:           comments,
		}
		goStruct.Fields = append(goStruct.Fields, goStructField)
	}
	return goStruct
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func goStructFromAlias(path string, info os.FileInfo, genDecl *ast.GenDecl, typeSpec *ast.TypeSpec, packageName string, alias *ast.Ident, file *ast.File) *GoStruct {
	// Get comments
	comments := make([]string, 0)
	if genDecl.Doc != nil && genDecl.Doc.List != nil && len(genDecl.Doc.List) > 0 {
		for _, comment := range genDecl.Doc.List {
			comments = append(comments, strings.TrimPrefix(comment.Text, "//"))
		}
	}

	usedStructType, usedStructPkgName := getUsedStructType(typeSpec.Type, packageName)
	tsType := fieldTypeToTypescriptType(typeSpec.Type, usedStructPkgName)

	goStruct := &GoStruct{
		Filepath:      filepath.ToSlash(path),
		Filename:      info.Name(),
		Name:          typeSpec.Name.Name,
		Package:       packageName,
		FormattedName: getTypePrefix(packageName) + typeSpec.Name.Name,
		Fields:        make([]*GoStructField, 0),
		Comments:      comments,
		AliasOf: &GoAlias{
			GoType:             alias.Name,
			TypescriptType:     tsType,
			UsedTypescriptType: fieldTypeToUsedTypescriptType(tsType),
			UsedStructType:     usedStructType,
		},
	}

	// Get declared values - useful for building enums or union types
	// e.g. const Something AliasType = "something"
	goStruct.AliasOf.DeclaredValues = make([]string, 0)
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			valueSpecType := fieldTypeString(valueSpec.Type)
			if len(valueSpec.Names) == 1 && valueSpec.Names[0].IsExported() && valueSpecType == typeSpec.Name.Name {
				for _, value := range valueSpec.Values {
					name, ok := value.(*ast.BasicLit)
					if !ok {
						continue
					}
					goStruct.AliasOf.DeclaredValues = append(goStruct.AliasOf.DeclaredValues, name.Value)
				}
			}
		}
	}
	return goStruct
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getUsedStructType returns the used struct type for a given type declaration.
// For example, if the type declaration is `map[string]models.User`, the used struct type is `models.User`.
// If the type declaration is `[]User`, the used struct type is `{packageName}.User`.
func getUsedStructType(expr ast.Expr, packageName string) (string, string) {
	usedStructType := fieldTypeToUsedStructType(expr)

	switch usedStructType {
	case "string", "bool", "byte", "uint", "uint8", "uint16", "uint32", "uint64", "int", "int8", "int16", "int32", "int64", "float", "float32", "float64":
		return "", ""
	case "__STRUCT__":
		return "", ""
	}

	if usedStructType != "__STRUCT__" && usedStructType != "" && !strings.Contains(usedStructType, ".") {
		usedStructType = packageName + "." + usedStructType
	}

	pkgName := strings.Split(usedStructType, ".")[0]

	return usedStructType, pkgName
}

// fieldTypeString returns the field type as a string.
// For example, if the field type is `[]*models.User`, the return value is `[]models.User`.
// If the field type is `[]InternalStruct`, the return value is `[]InternalStruct`.
func fieldTypeString(fieldType ast.Expr) string {
	switch t := fieldType.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		//return "*" + fieldTypeString(t.X)
		return fieldTypeString(t.X)
	case *ast.ArrayType:
		if fieldTypeString(t.Elt) == "byte" {
			return "string"
		}
		return "[]" + fieldTypeString(t.Elt)
	case *ast.MapType:
		return "map[" + fieldTypeString(t.Key) + "]" + fieldTypeString(t.Value)
	case *ast.SelectorExpr:
		return fieldTypeString(t.X) + "." + t.Sel.Name
	case *ast.StructType:
		return "__STRUCT__"
	default:
		return ""
	}
}

// fieldTypeToTypescriptType returns the field type as a string in TypeScript format.
// For example, if the field type is `[]*models.User`, the return value is `Array<Models_User>`.
func fieldTypeToTypescriptType(fieldType ast.Expr, usedStructPkgName string) string {
	switch t := fieldType.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "string"
		case "uint", "uint8", "uint16", "uint32", "uint64", "int", "int8", "int16", "int32", "int64", "float", "float32", "float64":
			return "number"
		case "bool":
			return "boolean"
		case "byte":
			return "string"
		case "time.Time":
			return "string"
		case "nil":
			return "null"
		default:
			return getTypePrefix(usedStructPkgName) + t.Name
		}
	case *ast.StarExpr:
		return fieldTypeToTypescriptType(t.X, usedStructPkgName)
	case *ast.ArrayType:
		if fieldTypeToTypescriptType(t.Elt, usedStructPkgName) == "byte" {
			return "string"
		}
		return "Array<" + fieldTypeToTypescriptType(t.Elt, usedStructPkgName) + ">"
	case *ast.MapType:
		return "Record<" + fieldTypeToTypescriptType(t.Key, usedStructPkgName) + ", " + fieldTypeToTypescriptType(t.Value, usedStructPkgName) + ">"
	case *ast.SelectorExpr:
		if t.Sel.Name == "Time" {
			return "string"
		}
		return getTypePrefix(usedStructPkgName) + t.Sel.Name
	case *ast.StructType:
		s := "{ "
		for _, field := range t.Fields.List {
			s += jsonFieldName(field) + ": " + fieldTypeToTypescriptType(field.Type, usedStructPkgName) + "; "
		}
		s += "}"
		return s
	default:
		return "any"
	}
}

func stringGoTypeToTypescriptType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "uint", "uint8", "uint16", "uint32", "uint64", "int", "int8", "int16", "int32", "int64", "float", "float32", "float64":
		return "number"
	case "nil":
		return "null"
	case "bool":
		return "boolean"
	case "time.Time":
		return "string"
	}

	if strings.HasPrefix(goType, "[]") {
		return "Array<" + stringGoTypeToTypescriptType(goType[2:]) + ">"
	}

	if strings.HasPrefix(goType, "*") {
		return stringGoTypeToTypescriptType(goType[1:])
	}

	if strings.HasPrefix(goType, "map[") {
		s := strings.TrimPrefix(goType, "map[")
		key := ""
		value := ""
		for i, c := range s {
			if c == ']' {
				key = s[:i]
				value = s[i+1:]
				break
			}
		}
		return "Record<" + stringGoTypeToTypescriptType(key) + ", " + stringGoTypeToTypescriptType(value) + ">"
	}

	if strings.Contains(goType, ".") {
		parts := strings.Split(goType, ".")
		return getTypePrefix(parts[0]) + parts[1]
	}

	return goType
}

func goTypeToTypescriptType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "uint", "uint8", "uint16", "uint32", "uint64", "int", "int8", "int16", "int32", "int64", "float", "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "nil":
		return "null"
	case "time.Time":
		return "string"
	default:
		return "unknown"
	}
}

// fieldTypeUnformattedString returns the field type as a string without formatting.
// For example, if the field type is `[]*models.User`, the return value is `models.User`.
// /!\ Caveat: this assumes that the map key is always a string.
func fieldTypeUnformattedString(fieldType ast.Expr) string {
	switch t := fieldType.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		//return "*" + fieldTypeString(t.X)
		return fieldTypeUnformattedString(t.X)
	case *ast.ArrayType:
		return fieldTypeUnformattedString(t.Elt)
	case *ast.MapType:
		return fieldTypeUnformattedString(t.Value)
	case *ast.SelectorExpr:
		return fieldTypeString(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

// fieldTypeToUsedStructType returns the used struct type for a given field type.
// For example, if the field type is `[]*models.User`, the return value is `models.User`.
func fieldTypeToUsedStructType(fieldType ast.Expr) string {
	switch t := fieldType.(type) {
	case *ast.StarExpr:
		return fieldTypeString(t.X)
	case *ast.ArrayType:
		return fieldTypeString(t.Elt)
	case *ast.MapType:
		return fieldTypeUnformattedString(t.Value)
	case *ast.SelectorExpr:
		return fieldTypeString(t)
	case *ast.Ident:
		return t.Name
	case *ast.StructType:
		return "__STRUCT__"
	default:
		return ""
	}
}

func jsonFieldName(field *ast.Field) string {
	if field.Tag != nil {
		tag := reflect.StructTag(strings.ReplaceAll(field.Tag.Value[1:len(field.Tag.Value)-1], "\\\"", "\""))
		jsonTag := tag.Get("json")
		if jsonTag != "" {
			jsonParts := strings.Split(jsonTag, ",")
			if jsonParts[0] == "-" {
				return ""
			}
			if jsonParts[0] != "" {
				return jsonParts[0]
			}
			return jsonParts[0]
		}
	}
	return field.Names[0].Name
}

func jsonFieldOmitEmpty(field *ast.Field) bool {
	if field.Tag != nil {
		tag := reflect.StructTag(strings.ReplaceAll(field.Tag.Value[1:len(field.Tag.Value)-1], "\\\"", "\""))
		jsonTag := tag.Get("json")
		if jsonTag != "" {
			jsonParts := strings.Split(jsonTag, ",")
			return len(jsonParts) > 1 && jsonParts[1] == "omitempty"
		}
	}
	return false
}

func isCustomStruct(goType string) bool {
	return goTypeToTypescriptType(goType) == "unknown"
}

var nameExceptions = map[string]string{"OAuth2": "oauth2"}

func convertGoToJSName(name string) string {
	if v, ok := nameExceptions[name]; ok {
		return v
	}

	startUppercase := make([]rune, 0, len(name))

	for _, c := range name {
		if c != '_' && !unicode.IsUpper(c) && !unicode.IsDigit(c) {
			break
		}

		startUppercase = append(startUppercase, c)
	}

	totalStartUppercase := len(startUppercase)

	// all uppercase eg. "JSON" -> "json"
	if len(name) == totalStartUppercase {
		return strings.ToLower(name)
	}

	// eg. "JSONField" -> "jsonField"
	if totalStartUppercase > 1 {
		return strings.ToLower(name[0:totalStartUppercase-1]) + name[totalStartUppercase-1:]
	}

	// eg. "GetField" -> "getField"
	if totalStartUppercase == 1 {
		return strings.ToLower(name[0:1]) + name[1:]
	}

	return name
}

// fieldTypeToUsedTypescriptType extracts the core TypeScript type from complex type expressions
// For example, if the type is Array<Models_User>, it returns Models_User
// If the type is Record<string, Models_User>, it returns Models_User
func fieldTypeToUsedTypescriptType(tsType string) string {
	// Handle arrays: Array<Type> -> Type
	if strings.HasPrefix(tsType, "Array<") && strings.HasSuffix(tsType, ">") {
		innerType := strings.TrimPrefix(strings.TrimSuffix(tsType, ">"), "Array<")
		return fieldTypeToUsedTypescriptType(innerType)
	}

	// Handle records: Record<Key, Value> -> Value
	if strings.HasPrefix(tsType, "Record<") && strings.HasSuffix(tsType, ">") {
		innerType := strings.TrimPrefix(strings.TrimSuffix(tsType, ">"), "Record<")
		// Find the comma that separates key and value
		commaIndex := -1
		bracketCount := 0
		for i, char := range innerType {
			if char == '<' {
				bracketCount++
			} else if char == '>' {
				bracketCount--
			} else if char == ',' && bracketCount == 0 {
				commaIndex = i
				break
			}
		}

		if commaIndex != -1 {
			valueType := strings.TrimSpace(innerType[commaIndex+1:])
			return fieldTypeToUsedTypescriptType(valueType)
		}
	}

	// Handle primitive types
	switch tsType {
	case "string", "number", "boolean", "any", "null", "undefined":
		return ""
	}

	return tsType
}
