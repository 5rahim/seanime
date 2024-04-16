package docs

import (
	"bufio"
	"bytes"
	"github.com/goccy/go-json"
	"os"
	"strings"
)

const (
	handlersDirPath = "../internal/handlers"
)

type (
	rawFile struct {
		Name    string
		Content []byte
	}

	Docs struct {
		RouteGroups []*RouteGroup `json:"routeGroups"`
	}

	RouteGroup struct {
		Filename string   `json:"filename"`
		Routes   []*Route `json:"routes"`
	}

	Route struct {
		Name              string               `json:"name"`
		Summary           string               `json:"summary"`
		Description       string               `json:"description"`
		Methods           []string             `json:"methods"`
		Endpoint          string               `json:"endpoint"`
		Params            []*Param             `json:"params"`
		RequestBodyFields []*RequestBodyFields `json:"requestBodyFields"`
		Returns           string               `json:"returns"`
	}

	Param struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Required    bool   `json:"required"`
		Description string `json:"description"`
	}
	RequestBodyFields struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
	}
)

func ParseRoutes(dir string) (docs *Docs) {

	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	rawFiles := make([]rawFile, 0)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasPrefix(file.Name(), "_") {
			continue
		}

		// Read the file
		content, err := os.ReadFile(handlersDirPath + "/" + file.Name())
		if err != nil {
			panic(err)
		}

		rawFiles = append(rawFiles, rawFile{
			Name:    file.Name(),
			Content: content,
		})

	}

	groups := make([]*RouteGroup, 0)

	for _, rawFile := range rawFiles {
		group := parseFile(rawFile.Content)
		group.Filename = rawFile.Name
		groups = append(groups, group)
	}

	docs = &Docs{
		RouteGroups: groups,
	}

	var retBytes []byte
	retBytes, err = json.Marshal(docs)
	if err != nil {
		panic(err)
	}

	// Write to json file
	err = os.WriteFile("routes.json", retBytes, 0644)
	if err != nil {
		panic(err)
	}

	return
}

type handlerInfo struct {
	CommentStart string
	Summary      string
	Descriptions []string
	Params       []string
	Route        string
	Returns      string
	FuncName     string
}

type handlerRequestBodyField struct {
	HandlerFuncName string
	JsonFieldName   string
	Type            string
}

func parseFile(content []byte) (ret *RouteGroup) {
	ret = &RouteGroup{
		Routes: make([]*Route, 0),
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	lineInfos := make([]*handlerInfo, 0)
	requestBodies := make([]*handlerRequestBodyField, 0)

	var currentLineInfo *handlerInfo
	var currentRequestBodyField *handlerRequestBodyField

	readRequestBody := false

	for scanner.Scan() {
		line := scanner.Bytes()

		// Start of a new route handler
		if bytes.HasPrefix(line, []byte("// Handle")) {

			if currentLineInfo == nil {
				currentLineInfo = &handlerInfo{}
			} else {
				lineInfos = append(lineInfos, currentLineInfo)
				currentLineInfo = &handlerInfo{}
			}

			endIndex := bytes.Index(line[3:], []byte(" "))
			if endIndex == -1 {
				currentLineInfo.CommentStart = string(line[3:])
			} else {
				currentLineInfo.CommentStart = string(line[3 : endIndex+3])
			}
		}
		if bytes.HasPrefix(line, []byte("//\t@route")) {
			currentLineInfo.Route = string(line[10:])
		}
		if bytes.HasPrefix(line, []byte("//\t@summary")) {
			currentLineInfo.Summary = string(line[12:])
		}
		if bytes.HasPrefix(line, []byte("//\t@desc")) {
			currentLineInfo.Descriptions = append(currentLineInfo.Descriptions, string(line[9:]))
		}
		if bytes.HasPrefix(line, []byte("//\t@param")) {
			currentLineInfo.Params = append(currentLineInfo.Params, string(line[10:]))
		}
		if bytes.HasPrefix(line, []byte("//\t@returns")) {
			currentLineInfo.Returns = string(line[12:])
		}
		if bytes.HasPrefix(line, []byte("func Handle")) {
			endIndex := bytes.Index(line, []byte("("))
			if currentLineInfo != nil {
				currentLineInfo.FuncName = string(line[5:endIndex])
			}
		}

		if bytes.Contains(line, []byte("type body struct {")) {
			readRequestBody = true
		}
		if readRequestBody {
			if bytes.Contains(line, []byte("}")) {
				readRequestBody = false
			}
			if bytes.Contains(line, []byte("`json:\"")) {
				lineFields := bytes.Fields(line)
				jsonFieldName := string(lineFields[0])
				if len(lineFields) > 2 {
					jsonFieldName = strings.Replace(string(lineFields[2]), "`json:\"", "", -1)
				}
				jsonFieldName = strings.Replace(jsonFieldName, ",omitempty", "", -1)
				jsonFieldName = strings.Replace(jsonFieldName, "\"`", "", -1)
				tType := string(lineFields[1])
				currentRequestBodyField = &handlerRequestBodyField{
					HandlerFuncName: currentLineInfo.FuncName,
					JsonFieldName:   jsonFieldName,
					Type:            tType,
				}
				requestBodies = append(requestBodies, currentRequestBodyField)
			}
		}

	}

	if currentLineInfo != nil {
		lineInfos = append(lineInfos, currentLineInfo)
	}

	for _, info := range lineInfos {

		methods := make([]string, 0)
		methodBrStart := strings.Index(info.Route, "[")
		methodBrEnd := strings.Index(info.Route, "]")
		if methodBrStart != -1 && methodBrEnd != -1 {
			methods = strings.Split(info.Route[methodBrStart+1:methodBrEnd], ",")
		}
		for i := range methods {
			methods[i] = strings.TrimSpace(methods[i])
		}
		endpoint := ""
		if methodBrStart != -1 {
			endpoint = info.Route[:methodBrStart]
		} else {
			endpoint = info.Route
		}
		endpoint = strings.TrimSpace(endpoint)

		reqBody := make([]*RequestBodyFields, 0)
		for _, bodyField := range requestBodies {
			if bodyField.HandlerFuncName == info.FuncName {
				reqBody = append(reqBody, &RequestBodyFields{
					Name:        bodyField.JsonFieldName,
					Type:        bodyField.Type,
					Description: "",
				})
			}
		}

		params := make([]*Param, 0)
		for _, param := range info.Params {
			// e.g. id - int - true - "The DB id of the rule"
			parts := strings.Split(param, " - ")
			if len(parts) != 4 {
				continue
			}
			required := parts[2] == "true"
			params = append(params, &Param{
				Name:        parts[0],
				Type:        parts[1],
				Required:    required,
				Description: strings.ReplaceAll(parts[3], "\"", ""),
			})
		}

		route := &Route{
			Name:              info.FuncName,
			Description:       strings.Join(info.Descriptions, " "),
			Endpoint:          endpoint,
			Summary:           info.Summary,
			Methods:           methods,
			Returns:           info.Returns,
			Params:            params,
			RequestBodyFields: reqBody,
		}
		ret.Routes = append(ret.Routes, route)
	}

	return
}
