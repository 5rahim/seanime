package codegen

import (
	"cmp"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	additionalStructNamesForHooks = []string{
		"discordrpc_presence.MangaActivity",
		"discordrpc_presence.AnimeActivity",
		"discordrpc_presence.LegacyAnimeActivity",
		"anilist.ListAnime",
		"anilist.ListManga",
		"anilist.MediaSort",
		"anilist.ListRecentAnime",
		"anilist.AnimeCollectionWithRelations",
		"onlinestream.Episode",
		"continuity.WatchHistoryItem",
		"continuity.WatchHistoryItemResponse",
		"continuity.UpdateWatchHistoryItemOptions",
		"continuity.WatchHistory",
		"torrent_client.Torrent",
	}
)

func GeneratePluginEventFile(inFilePath string, outDir string) {
	// Parse the input file
	file, err := parser.ParseFile(token.NewFileSet(), inFilePath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Create output directory if it doesn't exist
	_ = os.MkdirAll(outDir, os.ModePerm)

	const OutFileName = "plugin-events.ts"

	// Create output file
	f, err := os.Create(filepath.Join(outDir, OutFileName))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Write imports
	f.WriteString(`// This file is auto-generated. Do not edit.
	import { useWebsocketPluginMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
	import { useCallback } from "react"

`)

	// Extract client and server event types
	clientEvents := make([]string, 0)
	serverEvents := make([]string, 0)
	clientPayloads := make(map[string]string)
	serverPayloads := make(map[string]string)
	clientEventValues := make(map[string]string)
	serverEventValues := make(map[string]string)

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		// Find const declarations
		if genDecl.Tok == token.CONST {
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				if len(valueSpec.Names) == 1 && len(valueSpec.Values) == 1 {
					name := valueSpec.Names[0].Name
					if strings.HasPrefix(name, "Client") && strings.HasSuffix(name, "Event") {
						eventName := name[len("Client") : len(name)-len("Event")]
						// Get the string literal value for the enum
						if basicLit, ok := valueSpec.Values[0].(*ast.BasicLit); ok {
							eventValue := strings.Trim(basicLit.Value, "\"")
							clientEvents = append(clientEvents, eventName)
							// Get payload type name
							payloadType := name + "Payload"
							clientPayloads[eventName] = payloadType
							// Store the original string value
							clientEventValues[eventName] = eventValue
						}
					} else if strings.HasPrefix(name, "Server") && strings.HasSuffix(name, "Event") {
						eventName := name[len("Server") : len(name)-len("Event")]
						// Get the string literal value for the enum
						if basicLit, ok := valueSpec.Values[0].(*ast.BasicLit); ok {
							eventValue := strings.Trim(basicLit.Value, "\"")
							serverEvents = append(serverEvents, eventName)
							// Get payload type name
							payloadType := name + "Payload"
							serverPayloads[eventName] = payloadType
							// Store the original string value
							serverEventValues[eventName] = eventValue
						}
					}
				}
			}
		}
	}

	// Write enums
	f.WriteString("export enum PluginClientEvents {\n")
	for _, event := range clientEvents {
		enumName := toPascalCase(event)
		f.WriteString(fmt.Sprintf("    %s = \"%s\",\n", enumName, clientEventValues[event]))
	}
	f.WriteString("}\n\n")

	f.WriteString("export enum PluginServerEvents {\n")
	for _, event := range serverEvents {
		enumName := toPascalCase(event)
		f.WriteString(fmt.Sprintf("    %s = \"%s\",\n", enumName, serverEventValues[event]))
	}
	f.WriteString("}\n\n")

	// Write client to server section
	f.WriteString("/////////////////////////////////////////////////////////////////////////////////////\n")
	f.WriteString("// Client to server\n")
	f.WriteString("/////////////////////////////////////////////////////////////////////////////////////\n\n")

	// Write client event types and hooks
	for _, event := range clientEvents {
		// Get the payload type
		payloadType := clientPayloads[event]
		payloadFound := false

		// Find the payload type in the AST
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			if genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					if typeSpec.Name.Name == payloadType {
						payloadFound = true
						// Write the payload type
						f.WriteString(fmt.Sprintf("export type Plugin_Client_%sEventPayload = {\n", toPascalCase(event)))

						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							for _, field := range structType.Fields.List {
								if len(field.Names) > 0 {
									fieldName := jsonFieldName(field)
									fieldType := fieldTypeToTypescriptType(field.Type, "")
									f.WriteString(fmt.Sprintf("    %s: %s\n", fieldName, fieldType))
								}
							}
						}

						f.WriteString("}\n\n")

						// Write the hook
						hookName := fmt.Sprintf("usePluginSend%sEvent", toPascalCase(event))
						f.WriteString(fmt.Sprintf("export function %s() {\n", hookName))
						f.WriteString("    const { sendPluginMessage } = useWebsocketSender()\n")
						f.WriteString("\n")
						f.WriteString(fmt.Sprintf("    const send%sEvent = useCallback((payload: Plugin_Client_%sEventPayload, extensionID?: string) => {\n",
							toPascalCase(event), toPascalCase(event)))
						f.WriteString(fmt.Sprintf("        sendPluginMessage(PluginClientEvents.%s, payload, extensionID)\n",
							toPascalCase(event)))
						f.WriteString("    }, [])\n")
						f.WriteString("\n")
						f.WriteString("    return {\n")
						f.WriteString(fmt.Sprintf("        send%sEvent,\n", toPascalCase(event)))
						f.WriteString("    }\n")
						f.WriteString("}\n\n")
					}
				}
			}
		}

		// If payload type not found, write empty object type
		if !payloadFound {
			f.WriteString(fmt.Sprintf("export type Plugin_Client_%sEventPayload = {}\n\n", toPascalCase(event)))

			// Write the hook
			hookName := fmt.Sprintf("usePluginSend%sEvent", toPascalCase(event))
			f.WriteString(fmt.Sprintf("export function %s() {\n", hookName))
			f.WriteString("    const { sendPluginMessage } = useWebsocketSender()\n")
			f.WriteString("\n")
			f.WriteString(fmt.Sprintf("    const sendPlugin%sEvent = useCallback((payload: Plugin_Client_%sEventPayload, extensionID?: string) => {\n",
				toPascalCase(event), toPascalCase(event)))
			f.WriteString(fmt.Sprintf("        sendPluginMessage(PluginClientEvents.%s, payload, extensionID)\n",
				toPascalCase(event)))
			f.WriteString("    }, [])\n")
			f.WriteString("\n")
			f.WriteString("    return {\n")
			f.WriteString(fmt.Sprintf("        send%sEvent,\n", toPascalCase(event)))
			f.WriteString("    }\n")
			f.WriteString("}\n\n")
		}
	}

	// Write server to client section
	f.WriteString("/////////////////////////////////////////////////////////////////////////////////////\n")
	f.WriteString("// Server to client\n")
	f.WriteString("/////////////////////////////////////////////////////////////////////////////////////\n\n")

	// Write server event types and hooks
	for _, event := range serverEvents {
		// Get the payload type
		payloadType := serverPayloads[event]
		payloadFound := false

		// Find the payload type in the AST
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			if genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					if typeSpec.Name.Name == payloadType {
						payloadFound = true
						// Write the payload type
						f.WriteString(fmt.Sprintf("export type Plugin_Server_%sEventPayload = {\n", toPascalCase(event)))

						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							for _, field := range structType.Fields.List {
								if len(field.Names) > 0 {
									fieldName := jsonFieldName(field)
									fieldType := fieldTypeToTypescriptType(field.Type, "")
									f.WriteString(fmt.Sprintf("    %s: %s\n", fieldName, fieldType))
								}
							}
						}

						f.WriteString("}\n\n")

						// Write the hook
						hookName := fmt.Sprintf("usePluginListen%sEvent", toPascalCase(event))
						f.WriteString(fmt.Sprintf("export function %s(cb: (payload: Plugin_Server_%sEventPayload, extensionId: string) => void, extensionID: string) {\n",
							hookName, toPascalCase(event)))
						f.WriteString("    return useWebsocketPluginMessageListener<Plugin_Server_" + toPascalCase(event) + "EventPayload>({\n")
						f.WriteString("        extensionId: extensionID,\n")
						f.WriteString(fmt.Sprintf("        type: PluginServerEvents.%s,\n", toPascalCase(event)))
						f.WriteString("        onMessage: cb,\n")
						f.WriteString("    })\n")
						f.WriteString("}\n\n")
					}
				}
			}
		}

		// If payload type not found, write empty object type
		if !payloadFound {
			f.WriteString(fmt.Sprintf("export type Plugin_Server_%sEventPayload = {}\n\n", toPascalCase(event)))

			// Write the hook
			hookName := fmt.Sprintf("usePluginListen%sEvent", toPascalCase(event))
			f.WriteString(fmt.Sprintf("export function %s(cb: (payload: Plugin_Server_%sEventPayload, extensionId: string) => void, extensionID: string) {\n",
				hookName, toPascalCase(event)))
			f.WriteString("    return useWebsocketPluginMessageListener<Plugin_Server_" + toPascalCase(event) + "EventPayload>({\n")
			f.WriteString("        extensionId: extensionID,\n")
			f.WriteString(fmt.Sprintf("        type: PluginServerEvents.%s,\n", toPascalCase(event)))
			f.WriteString("        onMessage: cb,\n")
			f.WriteString("    })\n")
			f.WriteString("}\n\n")
		}
	}
}

var execptions = map[string]string{
	"playbackmanager": "PlaybackManager	",
}

func toPascalCase(s string) string {
	if exception, ok := execptions[s]; ok {
		return exception
	}
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = cases.Title(language.English, cases.NoLower).String(s)
	return strings.ReplaceAll(s, " ", "")
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type HookEventDefinition struct {
	Package  string    `json:"package"`
	GoStruct *GoStruct `json:"goStruct"`
}

func GeneratePluginHooksDefinitionFile(outDir string, publicStructsFilePath string, genOutDir string) {
	// Create output file
	f, err := os.Create(filepath.Join(outDir, "app.d.ts"))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	mdFile, err := os.Create(filepath.Join(genOutDir, "hooks.mdx"))
	if err != nil {
		panic(err)
	}
	defer mdFile.Close()

	goStructs := LoadPublicStructs(publicStructsFilePath)

	// e.g. map["models.User"]*GoStruct
	goStructsMap := make(map[string]*GoStruct)

	for _, goStruct := range goStructs {
		goStructsMap[goStruct.Package+"."+goStruct.Name] = goStruct
	}

	// Expand the structs with embedded structs
	for _, goStruct := range goStructs {
		for _, embeddedStructType := range goStruct.EmbeddedStructTypes {
			if embeddedStructType != "" {
				if usedStruct, ok := goStructsMap[embeddedStructType]; ok {
					for _, usedField := range usedStruct.Fields {
						goStruct.Fields = append(goStruct.Fields, usedField)
					}
				}
			}
		}
	}

	// Key = package
	eventGoStructsMap := make(map[string][]*GoStruct)
	for _, goStruct := range goStructs {
		if goStruct.Filename == "hook_events.go" {
			if _, ok := eventGoStructsMap[goStruct.Package]; !ok {
				eventGoStructsMap[goStruct.Package] = make([]*GoStruct, 0)
			}
			eventGoStructsMap[goStruct.Package] = append(eventGoStructsMap[goStruct.Package], goStruct)
		}
	}

	// Create `hooks.json`
	hookEventDefinitions := make([]*HookEventDefinition, 0)
	for _, goStruct := range goStructs {
		if goStruct.Filename == "hook_events.go" {
			hookEventDefinitions = append(hookEventDefinitions, &HookEventDefinition{
				Package:  goStruct.Package,
				GoStruct: goStruct,
			})
		}
	}
	jsonFile, err := os.Create(filepath.Join(genOutDir, "hooks.json"))
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(hookEventDefinitions); err != nil {
		fmt.Println("Error:", err)
		return
	}

	////////////////////////////////////////////////////
	// Write `app.d.ts`
	// Write namespace declaration
	////////////////////////////////////////////////////
	f.WriteString("declare namespace $app {\n")

	packageNames := make([]string, 0)
	for packageName := range eventGoStructsMap {
		packageNames = append(packageNames, packageName)
	}
	slices.Sort(packageNames)

	//////////////////////////////////////////////////////////
	// Get referenced structs so we can write them at the end
	//////////////////////////////////////////////////////////
	sharedStructs := make([]*GoStruct, 0)
	otherStructs := make([]*GoStruct, 0)

	// Go through all the event structs' fields, and get the types that are structs
	sharedStructsMap := make(map[string]*GoStruct)
	for _, goStructs := range eventGoStructsMap {
		for _, goStruct := range goStructs {
			for _, field := range goStruct.Fields {
				if isCustomStruct(field.GoType) {
					if _, ok := sharedStructsMap[field.GoType]; !ok && goStructsMap[field.UsedStructType] != nil {
						sharedStructsMap[field.UsedStructType] = goStructsMap[field.UsedStructType]
					}
				}
			}
		}
	}

	// Add additional structs to otherStructs
	for _, structName := range additionalStructNamesForHooks {
		if _, ok := sharedStructsMap[structName]; !ok {
			sharedStructsMap[structName] = goStructsMap[structName]
		}
	}

	for _, goStruct := range sharedStructsMap {
		//fmt.Println(goStruct.FormattedName)
		if goStruct.Package != "" {
			sharedStructs = append(sharedStructs, goStruct)
		}
	}

	referencedStructsMap, ok := getReferencedStructsRecursively(sharedStructs, otherStructs, goStructsMap)
	if !ok {
		panic("Failed to get referenced structs")
	}

	for _, packageName := range packageNames {
		writePackageEventGoStructs(f, packageName, eventGoStructsMap[packageName], goStructsMap)
	}

	f.WriteString("    ///////////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
	f.WriteString("    ///////////////////////////////////////////////////////////////////////////////////////////////////////////////\n")
	f.WriteString("    ///////////////////////////////////////////////////////////////////////////////////////////////////////////////\n\n")

	referencedStructs := make([]*GoStruct, 0)
	for _, goStruct := range referencedStructsMap {
		//fmt.Println(goStruct.FormattedName)
		referencedStructs = append(referencedStructs, goStruct)
	}
	slices.SortFunc(referencedStructs, func(a, b *GoStruct) int {
		return strings.Compare(a.FormattedName, b.FormattedName)
	})

	// Write the shared structs at the end
	for _, goStruct := range referencedStructs {
		if goStruct.Package != "" {
			writeEventTypescriptType(f, goStruct, make(map[string]*GoStruct))
		}
	}

	f.WriteString("}\n")

	// Generate markdown documentation
	writeMarkdownFile(mdFile, hookEventDefinitions, referencedStructsMap, referencedStructs)

}

func writePackageEventGoStructs(f *os.File, packageName string, goStructs []*GoStruct, allGoStructs map[string]*GoStruct) {
	// Header comment block
	f.WriteString(fmt.Sprintf("\n    /**\n     * @package %s\n     */\n\n", packageName))

	// Declare the hook functions
	for _, goStruct := range goStructs {
		// Write comments
		comments := ""
		comments += fmt.Sprintf("\n     * @event %s\n", goStruct.Name)
		comments += fmt.Sprintf("     * @file %s\n", strings.TrimPrefix(goStruct.Filepath, "../"))

		shouldAddPreventDefault := false

		if len(goStruct.Comments) > 0 {
			comments += fmt.Sprintf("     * @description\n")
		}
		for _, comment := range goStruct.Comments {
			if strings.Contains(strings.ToLower(comment), "prevent default") {
				shouldAddPreventDefault = true
			}
			comments += fmt.Sprintf("     * %s\n", strings.TrimSpace(comment))
		}
		f.WriteString(fmt.Sprintf("    /**%s     */\n", comments))

		//////// Write hook function
		f.WriteString(fmt.Sprintf("    function on%s(cb: (event: %s) => void): void;\n\n", strings.TrimSuffix(goStruct.Name, "Event"), goStruct.Name))

		/////// Write event interface
		f.WriteString(fmt.Sprintf("    interface %s {\n", goStruct.Name))
		f.WriteString(fmt.Sprintf("        next(): void;\n\n"))
		if shouldAddPreventDefault {
			f.WriteString(fmt.Sprintf("        preventDefault(): void;\n\n"))
		}
		// Write the fields
		for _, field := range goStruct.Fields {
			if field.Name == "next" || field.Name == "preventDefault" || field.Name == "DefaultPrevented" {
				continue
			}
			if field.JsonName == "" {
				continue
			}
			// Field type
			fieldNameSuffix := ""
			if !field.Required {
				fieldNameSuffix = "?"
			}

			if len(field.Comments) > 0 {
				f.WriteString(fmt.Sprintf("    /**\n"))
				for _, cmt := range field.Comments {
					f.WriteString(fmt.Sprintf("     * %s\n", strings.TrimSpace(cmt)))
				}
				f.WriteString(fmt.Sprintf("     */\n"))
			}

			typeText := field.TypescriptType

			f.WriteString(fmt.Sprintf("        %s%s: %s;\n", field.JsonName, fieldNameSuffix, typeText))
		}
		f.WriteString(fmt.Sprintf("    }\n\n"))

	}
}

func writeEventTypescriptType(f *os.File, goStruct *GoStruct, writtenTypes map[string]*GoStruct) {
	f.WriteString("    /**\n")
	f.WriteString(fmt.Sprintf("     * - Filepath: %s\n", strings.TrimPrefix(goStruct.Filepath, "../")))
	if len(goStruct.Comments) > 0 {
		f.WriteString(fmt.Sprintf("     * @description\n"))
		for _, cmt := range goStruct.Comments {
			f.WriteString(fmt.Sprintf("     *  %s\n", strings.TrimSpace(cmt)))
		}
	}
	f.WriteString("     */\n")

	if len(goStruct.Fields) > 0 {
		f.WriteString(fmt.Sprintf("    interface %s {\n", goStruct.FormattedName))
		for _, field := range goStruct.Fields {
			fieldNameSuffix := ""
			if !field.Required {
				fieldNameSuffix = "?"
			}
			if field.JsonName == "" {
				continue
			}

			if len(field.Comments) > 0 {
				f.WriteString(fmt.Sprintf("        /**\n"))
				for _, cmt := range field.Comments {
					f.WriteString(fmt.Sprintf("         * %s\n", strings.TrimSpace(cmt)))
				}
				f.WriteString(fmt.Sprintf("         */\n"))
			}

			typeText := field.TypescriptType
			if typeText == "Habari_Metadata" {
				typeText = "$habari.Metadata"
			}

			f.WriteString(fmt.Sprintf("        %s%s: %s;\n", field.JsonName, fieldNameSuffix, typeText))
		}
		f.WriteString("    }\n\n")
	}

	if goStruct.AliasOf != nil {
		if goStruct.AliasOf.DeclaredValues != nil && len(goStruct.AliasOf.DeclaredValues) > 0 {
			union := ""
			if len(goStruct.AliasOf.DeclaredValues) > 5 {
				union = strings.Join(goStruct.AliasOf.DeclaredValues, " |\n    ")
			} else {
				union = strings.Join(goStruct.AliasOf.DeclaredValues, " | ")
			}
			f.WriteString(fmt.Sprintf("    export type %s = %s;\n\n", goStruct.FormattedName, union))
		} else {
			f.WriteString(fmt.Sprintf("    export type %s = %s;\n\n", goStruct.FormattedName, goStruct.AliasOf.TypescriptType))
		}
	}

	// Add the struct to the written types
	writtenTypes[goStruct.Package+"."+goStruct.Name] = goStruct
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// writeMarkdownFile generates a well-formatted Markdown documentation for hooks
func writeMarkdownFile(mdFile *os.File, hookEventDefinitions []*HookEventDefinition, referencedStructsMap map[string]*GoStruct, referencedStructs []*GoStruct) {

	mdFile.WriteString("---\n")
	mdFile.WriteString("title: Hooks\n")
	mdFile.WriteString("description: How to use hooks\n")
	mdFile.WriteString("---")
	mdFile.WriteString("\n\n")

	// Group hooks by package
	packageHooks := make(map[string][]*HookEventDefinition)
	for _, hook := range hookEventDefinitions {
		packageHooks[hook.Package] = append(packageHooks[hook.Package], hook)
	}

	// Sort packages alphabetically
	packageNames := make([]string, 0, len(packageHooks))
	for pkg := range packageHooks {
		packageNames = append(packageNames, pkg)
	}
	slices.Sort(packageNames)

	// Write each package section
	for _, pkg := range packageNames {
		hooks := packageHooks[pkg]

		mdFile.WriteString(fmt.Sprintf("<a id=\"%s\"></a>\n", pkg))
		mdFile.WriteString(fmt.Sprintf("# %s\n\n", toPascalCase(pkg)))

		// Write each hook in the package
		for _, hook := range hooks {
			goStruct := hook.GoStruct
			eventName := goStruct.Name
			hookName := fmt.Sprintf("on%s", strings.TrimSuffix(eventName, "Event"))

			mdFile.WriteString(fmt.Sprintf("<a id=\"on%s\"></a>\n", strings.ToLower(strings.TrimSuffix(eventName, "Event"))))
			mdFile.WriteString(fmt.Sprintf("## %s\n\n", hookName))

			// Write description
			if len(goStruct.Comments) > 0 {
				for _, comment := range goStruct.Comments {
					mdFile.WriteString(fmt.Sprintf("%s\n", strings.TrimSpace(comment)))
				}
				mdFile.WriteString("\n")
			}

			// Check if it has preventDefault
			hasPreventDefault := false
			for _, comment := range goStruct.Comments {
				if strings.Contains(strings.ToLower(comment), "prevent default") {
					hasPreventDefault = true
					break
				}
			}

			if hasPreventDefault {
				mdFile.WriteString("**Can prevent default:** Yes\n\n")
			} else {
				mdFile.WriteString("**Can prevent default:** No\n\n")
			}

			// Write event interface
			mdFile.WriteString("**Event Interface:**\n\n")
			mdFile.WriteString("```typescript\n")
			mdFile.WriteString(fmt.Sprintf("interface %s {\n", eventName))
			mdFile.WriteString("    next();\n")
			if hasPreventDefault {
				mdFile.WriteString("    preventDefault();\n")
			}

			// Write fields
			for _, field := range goStruct.Fields {
				if field.Name == "next" || field.Name == "preventDefault" || field.Name == "DefaultPrevented" {
					continue
				}
				if field.JsonName == "" {
					continue
				}

				fieldNameSuffix := ""
				if !field.Required {
					fieldNameSuffix = "?"
				}

				// Add comments if available
				if len(field.Comments) > 0 {
					mdFile.WriteString("\n    /**\n")
					for _, comment := range field.Comments {
						mdFile.WriteString(fmt.Sprintf("     * %s\n", strings.TrimSpace(comment)))
					}
					mdFile.WriteString("     */\n")
				}

				mdFile.WriteString(fmt.Sprintf("    %s%s: %s;\n", field.JsonName, fieldNameSuffix, field.TypescriptType))
			}

			mdFile.WriteString("}\n")
			mdFile.WriteString("```\n\n")

			referenced := make([]*GoStruct, 0)
			for _, field := range goStruct.Fields {
				if !isCustomStruct(field.GoType) {
					continue
				}
				goStruct, ok := referencedStructsMap[field.UsedStructType]
				if !ok {
					continue
				}
				referenced = append(referenced, goStruct)
			}

			// Add a list of referenced structs links
			if len(referenced) > 0 {
				mdFile.WriteString("**Event types:**\n\n")
			}
			for _, goStruct := range referenced {
				mdFile.WriteString(fmt.Sprintf("- [%s](#%s)\n", goStruct.FormattedName, goStruct.FormattedName))
			}
			mdFile.WriteString("\n")

			// Add example usage
			mdFile.WriteString("**Example:**\n\n")
			mdFile.WriteString("```typescript\n")
			mdFile.WriteString(fmt.Sprintf("$app.%s((e) => {\n", hookName))

			// Generate example code based on fields
			for _, field := range goStruct.Fields {
				if field.Name == "next" || field.Name == "preventDefault" || field.Name == "DefaultPrevented" {
					continue
				}

				mdFile.WriteString(fmt.Sprintf("    // console.log(e.%s);\n", field.JsonName))
			}

			if hasPreventDefault {
				mdFile.WriteString("\n    // Prevent default behavior if needed\n")
				mdFile.WriteString("    // e.preventDefault();\n")
			}

			mdFile.WriteString("    \n    e.next();\n")
			mdFile.WriteString("});\n")
			mdFile.WriteString("```\n\n")

			// Add separator between hooks
			mdFile.WriteString("---\n\n")
		}
	}

	// Write the referenced structs
	if len(referencedStructs) > 0 {
		mdFile.WriteString("\n# Referenced Types\n\n")
	}
	for _, goStruct := range referencedStructs {

		mdFile.WriteString(fmt.Sprintf("#### %s\n\n", goStruct.FormattedName))
		mdFile.WriteString(fmt.Sprintf("<div id=\"%s\"></div>\n\n", goStruct.FormattedName))
		mdFile.WriteString(fmt.Sprintf("**Filepath:** `%s`\n\n", strings.TrimPrefix(goStruct.Filepath, "../")))

		if len(goStruct.Fields) > 0 {
			mdFile.WriteString("**Fields:**\n\n")

			mdFile.WriteString("<Table>\n")
			mdFile.WriteString("<TableCaption>Fields</TableCaption>\n")
			mdFile.WriteString("<TableHeader>\n")
			mdFile.WriteString("<TableRow>\n")
			mdFile.WriteString("<TableHead>Property</TableHead>\n")
			mdFile.WriteString("<TableHead>Type</TableHead>\n")
			mdFile.WriteString("<TableHead>Description</TableHead>\n")
			mdFile.WriteString("</TableRow>\n")
			mdFile.WriteString("</TableHeader>\n")
			mdFile.WriteString("<TableBody>\n")
			for _, field := range goStruct.Fields {
				mdFile.WriteString(fmt.Sprintf("<TableRow>\n"))
				mdFile.WriteString(fmt.Sprintf("<TableCell className=\"py-1 px-2 max-w-[200px] break-all\">%s</TableCell>\n", field.JsonName))

				typeContainsReference := false
				if field.UsedStructType != "" && isCustomStruct(field.UsedStructType) {
					typeContainsReference = true
				}
				if typeContainsReference {
					link := fmt.Sprintf("<a href=\"#%s\">`%s`</a>", field.UsedTypescriptType, field.TypescriptType)
					mdFile.WriteString(fmt.Sprintf("<TableCell className=\"py-1 px-2 break-all\">%s</TableCell>\n", link))
				} else {
					mdFile.WriteString(fmt.Sprintf("<TableCell className=\"py-1 px-2 break-all\">`%s`</TableCell>\n", field.TypescriptType))
				}
				mdFile.WriteString(fmt.Sprintf("<TableCell className=\"py-1 px-2 max-w-[200px] break-all\">%s</TableCell>\n", cmp.Or(strings.Join(field.Comments, "\n"), "-")))
				mdFile.WriteString("</TableRow>\n")
			}
			mdFile.WriteString("</TableBody>\n")
			mdFile.WriteString("</Table>\n")

		}

		if goStruct.AliasOf != nil {
			if goStruct.AliasOf.DeclaredValues != nil && len(goStruct.AliasOf.DeclaredValues) > 0 {
				union := ""
				if len(goStruct.AliasOf.DeclaredValues) > 5 {
					union = strings.Join(goStruct.AliasOf.DeclaredValues, " |\n    ")
				} else {
					union = strings.Join(goStruct.AliasOf.DeclaredValues, " | ")
				}
				mdFile.WriteString(fmt.Sprintf("`%s`\n\n", union))
			} else {
				mdFile.WriteString(fmt.Sprintf("`%s`\n\n", goStruct.AliasOf.TypescriptType))
			}
		}

		mdFile.WriteString("\n")
	}
}
