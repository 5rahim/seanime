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
	outFile.WriteString(`// This file is auto-generated. Do not edit.
import { useCallback } from "react"
import { useWebsocketPluginMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"

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
	outFile.WriteString("export enum PluginClientEvents {\n")
	for _, event := range clientEvents {
		enumName := toPascalCase(event)
		outFile.WriteString(fmt.Sprintf("    %s = \"%s\",\n", enumName, clientEventValues[event]))
	}
	outFile.WriteString("}\n\n")

	outFile.WriteString("export enum PluginServerEvents {\n")
	for _, event := range serverEvents {
		enumName := toPascalCase(event)
		outFile.WriteString(fmt.Sprintf("    %s = \"%s\",\n", enumName, serverEventValues[event]))
	}
	outFile.WriteString("}\n\n")

	// Write client to server section
	outFile.WriteString("/////////////////////////////////////////////////////////////////////////////////////\n")
	outFile.WriteString("// Client to server\n")
	outFile.WriteString("/////////////////////////////////////////////////////////////////////////////////////\n\n")

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
						outFile.WriteString(fmt.Sprintf("export type Plugin_Client_%sEventPayload = {\n", toPascalCase(event)))

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
						hookName := fmt.Sprintf("usePluginSend%sEvent", toPascalCase(event))
						outFile.WriteString(fmt.Sprintf("export function %s() {\n", hookName))
						outFile.WriteString("    const { sendPluginMessage } = useWebsocketSender()\n")
						outFile.WriteString("\n")
						outFile.WriteString(fmt.Sprintf("    const send%sEvent = useCallback((payload: Plugin_Client_%sEventPayload, extensionID?: string) => {\n",
							toPascalCase(event), toPascalCase(event)))
						outFile.WriteString(fmt.Sprintf("        sendPluginMessage(PluginClientEvents.%s, payload, extensionID)\n",
							toPascalCase(event)))
						outFile.WriteString("    }, [])\n")
						outFile.WriteString("\n")
						outFile.WriteString("    return {\n")
						outFile.WriteString(fmt.Sprintf("        send%sEvent\n", toPascalCase(event)))
						outFile.WriteString("    }\n")
						outFile.WriteString("}\n\n")
					}
				}
			}
		}

		// If payload type not found, write empty object type
		if !payloadFound {
			outFile.WriteString(fmt.Sprintf("export type Plugin_Client_%sEventPayload = {}\n\n", toPascalCase(event)))

			// Write the hook
			hookName := fmt.Sprintf("usePluginSend%sEvent", toPascalCase(event))
			outFile.WriteString(fmt.Sprintf("export function %s() {\n", hookName))
			outFile.WriteString("    const { sendPluginMessage } = useWebsocketSender()\n")
			outFile.WriteString("\n")
			outFile.WriteString(fmt.Sprintf("    const sendPlugin%sEvent = useCallback((payload: Plugin_Client_%sEventPayload, extensionID?: string) => {\n",
				toPascalCase(event), toPascalCase(event)))
			outFile.WriteString(fmt.Sprintf("        sendPluginMessage(PluginClientEvents.%s, payload, extensionID)\n",
				toPascalCase(event)))
			outFile.WriteString("    }, [])\n")
			outFile.WriteString("\n")
			outFile.WriteString("    return {\n")
			outFile.WriteString(fmt.Sprintf("        send%sEvent\n", toPascalCase(event)))
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
						outFile.WriteString(fmt.Sprintf("export type Plugin_Server_%sEventPayload = {\n", toPascalCase(event)))

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
						hookName := fmt.Sprintf("usePluginListen%sEvent", toPascalCase(event))
						outFile.WriteString(fmt.Sprintf("export function %s(cb: (payload: Plugin_Server_%sEventPayload, extensionId: string) => void, extensionID: string) {\n",
							hookName, toPascalCase(event)))
						outFile.WriteString("    return useWebsocketPluginMessageListener<Plugin_Server_" + toPascalCase(event) + "EventPayload>({\n")
						outFile.WriteString("        extensionId: extensionID,\n")
						outFile.WriteString(fmt.Sprintf("        type: PluginServerEvents.%s,\n", toPascalCase(event)))
						outFile.WriteString("        onMessage: cb,\n")
						outFile.WriteString("    })\n")
						outFile.WriteString("}\n\n")
					}
				}
			}
		}

		// If payload type not found, write empty object type
		if !payloadFound {
			outFile.WriteString(fmt.Sprintf("export type Plugin_Server_%sEventPayload = {}\n\n", toPascalCase(event)))

			// Write the hook
			hookName := fmt.Sprintf("usePluginListen%sEvent", toPascalCase(event))
			outFile.WriteString(fmt.Sprintf("export function %s(cb: (payload: Plugin_Server_%sEventPayload, extensionId: string) => void, extensionID: string) {\n",
				hookName, toPascalCase(event)))
			outFile.WriteString("    return useWebsocketPluginMessageListener<Plugin_Server_" + toPascalCase(event) + "EventPayload>({\n")
			outFile.WriteString("        extensionId: extensionID,\n")
			outFile.WriteString(fmt.Sprintf("        type: PluginServerEvents.%s,\n", toPascalCase(event)))
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
