package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	outFile, err := os.Create(filepath.Join(outDir, OutFileName))
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// Write imports
	outFile.WriteString(`import { useCallback } from "react"
import { useWebsocketPluginMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"

`)

	// Extract client and server event types
	clientEvents := make([]string, 0)
	serverEvents := make([]string, 0)
	clientPayloads := make(map[string]string)
	serverPayloads := make(map[string]string)

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
						if basicLit, ok := valueSpec.Values[0].(*ast.BasicLit); ok {
							eventName := strings.Trim(basicLit.Value, "\"")
							clientEvents = append(clientEvents, eventName)
							// Get payload type name
							payloadType := name + "Payload"
							clientPayloads[eventName] = payloadType
						}
					} else if strings.HasPrefix(name, "Server") && strings.HasSuffix(name, "Event") {
						if basicLit, ok := valueSpec.Values[0].(*ast.BasicLit); ok {
							eventName := strings.Trim(basicLit.Value, "\"")
							serverEvents = append(serverEvents, eventName)
							// Get payload type name
							payloadType := name + "Payload"
							serverPayloads[eventName] = payloadType
						}
					}
				}
			}
		}
	}

	// Write enums
	outFile.WriteString("export enum PluginClientEvents {\n")
	for _, event := range clientEvents {
		parts := strings.Split(event, ":")
		if len(parts) == 2 {
			enumName := toPascalCase(parts[0]) + toPascalCase(parts[1])
			outFile.WriteString(fmt.Sprintf("    %s = \"%s\",\n", enumName, event))
		}
	}
	outFile.WriteString("}\n\n")

	outFile.WriteString("export enum PluginServerEvents {\n")
	for _, event := range serverEvents {
		parts := strings.Split(event, ":")
		if len(parts) == 2 {
			enumName := toPascalCase(parts[0]) + toPascalCase(parts[1])
			outFile.WriteString(fmt.Sprintf("    %s = \"%s\",\n", enumName, event))
		}
	}
	outFile.WriteString("}\n\n")

	// Write client to server section
	outFile.WriteString("/////////////////////////////////////////////////////////////////////////////////////\n")
	outFile.WriteString("// Client to server\n")
	outFile.WriteString("/////////////////////////////////////////////////////////////////////////////////////\n\n")

	// Write client event types and hooks
	for _, event := range clientEvents {
		parts := strings.Split(event, ":")
		if len(parts) != 2 {
			continue
		}

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
						outFile.WriteString(fmt.Sprintf("export type Plugin_Client_%s%sEventPayload = {\n", toPascalCase(parts[0]), toPascalCase(parts[1])))

						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							for _, field := range structType.Fields.List {
								if len(field.Names) > 0 {
									fieldName := jsonFieldName(field)
									fieldType := fieldTypeToTypescriptType(field.Type, "")
									outFile.WriteString(fmt.Sprintf("    %s: %s\n", fieldName, fieldType))
								}
							}
						}

						outFile.WriteString("}\n\n")

						// Write the hook
						hookName := fmt.Sprintf("usePluginSend%s%sEvent", toPascalCase(parts[0]), toPascalCase(parts[1]))
						outFile.WriteString(fmt.Sprintf("export function %s() {\n", hookName))
						outFile.WriteString("    const { sendPluginMessage } = useWebsocketSender()\n")
						outFile.WriteString("\n")
						outFile.WriteString(fmt.Sprintf("    const send%s%sEvent = useCallback((payload: Plugin_Client_%s%sEventPayload, extensionID?: string) => {\n",
							toPascalCase(parts[0]), toPascalCase(parts[1]), toPascalCase(parts[0]), toPascalCase(parts[1])))
						outFile.WriteString(fmt.Sprintf("        sendPluginMessage(PluginClientEvents.%s%s, payload, extensionID)\n",
							toPascalCase(parts[0]), toPascalCase(parts[1])))
						outFile.WriteString("    }, [])\n")
						outFile.WriteString("\n")
						outFile.WriteString("    return {\n")
						outFile.WriteString(fmt.Sprintf("        send%s%sEvent\n", toPascalCase(parts[0]), toPascalCase(parts[1])))
						outFile.WriteString("    }\n")
						outFile.WriteString("}\n\n")
					}
				}
			}
		}

		// If payload type not found, write empty object type
		if !payloadFound {
			outFile.WriteString(fmt.Sprintf("export type Plugin_Client_%s%sEventPayload = {}\n\n", toPascalCase(parts[0]), toPascalCase(parts[1])))

			// Write the hook
			hookName := fmt.Sprintf("usePluginSend%s%sEvent", toPascalCase(parts[0]), toPascalCase(parts[1]))
			outFile.WriteString(fmt.Sprintf("export function %s() {\n", hookName))
			outFile.WriteString("    const { sendPluginMessage } = useWebsocketSender()\n")
			outFile.WriteString("\n")
			outFile.WriteString(fmt.Sprintf("    const sendPlugin%s%sEvent = useCallback((payload: Plugin_Client_%s%sEventPayload, extensionID?: string) => {\n",
				toPascalCase(parts[0]), toPascalCase(parts[1]), toPascalCase(parts[0]), toPascalCase(parts[1])))
			outFile.WriteString(fmt.Sprintf("        sendPluginMessage(PluginClientEvents.%s%s, payload, extensionID)\n",
				toPascalCase(parts[0]), toPascalCase(parts[1])))
			outFile.WriteString("    }, [])\n")
			outFile.WriteString("\n")
			outFile.WriteString("    return {\n")
			outFile.WriteString(fmt.Sprintf("        send%s%sEvent\n", toPascalCase(parts[0]), toPascalCase(parts[1])))
			outFile.WriteString("    }\n")
			outFile.WriteString("}\n\n")
		}
	}

	// Write server to client section
	outFile.WriteString("/////////////////////////////////////////////////////////////////////////////////////\n")
	outFile.WriteString("// Server to client\n")
	outFile.WriteString("/////////////////////////////////////////////////////////////////////////////////////\n\n")

	// Write server event types and hooks
	for _, event := range serverEvents {
		parts := strings.Split(event, ":")
		if len(parts) != 2 {
			continue
		}

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
						outFile.WriteString(fmt.Sprintf("export type Plugin_Server_%s%sEventPayload = {\n", toPascalCase(parts[0]), toPascalCase(parts[1])))

						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							for _, field := range structType.Fields.List {
								if len(field.Names) > 0 {
									fieldName := jsonFieldName(field)
									fieldType := fieldTypeToTypescriptType(field.Type, "")
									outFile.WriteString(fmt.Sprintf("    %s: %s\n", fieldName, fieldType))
								}
							}
						}

						outFile.WriteString("}\n\n")

						// Write the hook
						hookName := fmt.Sprintf("usePluginListen%s%sEvent", toPascalCase(parts[0]), toPascalCase(parts[1]))
						outFile.WriteString(fmt.Sprintf("export function %s(cb: (payload: Plugin_Server_%s%sEventPayload) => void, extensionID: string) {\n",
							hookName, toPascalCase(parts[0]), toPascalCase(parts[1])))
						outFile.WriteString("    return useWebsocketPluginMessageListener<Plugin_Server_" + toPascalCase(parts[0]) + toPascalCase(parts[1]) + "EventPayload>({\n")
						outFile.WriteString("        extensionId: extensionID,\n")
						outFile.WriteString(fmt.Sprintf("        type: PluginServerEvents.%s%s,\n", toPascalCase(parts[0]), toPascalCase(parts[1])))
						outFile.WriteString("        onMessage: cb,\n")
						outFile.WriteString("    })\n")
						outFile.WriteString("}\n\n")
					}
				}
			}
		}

		// If payload type not found, write empty object type
		if !payloadFound {
			outFile.WriteString(fmt.Sprintf("export type Plugin_Server_%s%sEventPayload = {}\n\n", toPascalCase(parts[0]), toPascalCase(parts[1])))

			// Write the hook
			hookName := fmt.Sprintf("usePluginListen%s%sEvent", toPascalCase(parts[0]), toPascalCase(parts[1]))
			outFile.WriteString(fmt.Sprintf("export function %s(cb: (payload: Plugin_Server_%s%sEventPayload) => void, extensionID: string) {\n",
				hookName, toPascalCase(parts[0]), toPascalCase(parts[1])))
			outFile.WriteString("    return useWebsocketPluginMessageListener<Plugin_Server_" + toPascalCase(parts[0]) + toPascalCase(parts[1]) + "EventPayload>({\n")
			outFile.WriteString("        extensionId: extensionID,\n")
			outFile.WriteString(fmt.Sprintf("        type: PluginServerEvents.%s%s,\n", toPascalCase(parts[0]), toPascalCase(parts[1])))
			outFile.WriteString("        onMessage: cb,\n")
			outFile.WriteString("    })\n")
			outFile.WriteString("}\n\n")
		}
	}
}

func toPascalCase(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	s = cases.Title(language.English, cases.NoLower).String(s)
	return strings.ReplaceAll(s, " ", "")
}
