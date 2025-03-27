package plugin_ui

import (
	"errors"
	"fmt"

	"github.com/dop251/goja"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

func (c *ComponentManager) renderComponents(renderFunc func(goja.FunctionCall) goja.Value) (interface{}, error) {
	if renderFunc == nil {
		return nil, errors.New("render function is not set")
	}

	// Get new components
	newComponents := c.getComponentsData(renderFunc)

	// If we have previous components, perform diffing
	if c.lastRenderedComponents != nil {
		newComponents = c.componentDiff(c.lastRenderedComponents, newComponents)
	}

	// Store the new components for next render
	c.lastRenderedComponents = newComponents

	return newComponents, nil
}

// getComponentsData calls the render function and returns the current state of the component tree
func (c *ComponentManager) getComponentsData(renderFunc func(goja.FunctionCall) goja.Value) interface{} {
	// Call the render function
	value := renderFunc(goja.FunctionCall{})

	// Convert the value to a JSON string
	v, err := json.Marshal(value)
	if err != nil {
		return nil
	}

	var ret interface{}
	err = json.Unmarshal(v, &ret)
	if err != nil {
		return nil
	}

	return ret
}

////

type ComponentProp struct {
	Name             string                        // e.g. "label"
	Type             string                        // e.g. "string"
	Default          interface{}                   // Is set if the prop is not provided, if not set and required is false, the prop will not be included in the component
	Required         bool                          // If true an no default value is provided, the component will throw a type error
	Validate         func(value interface{}) error // Optional validation function
	OptionalFirstArg bool                          // If true, it can be the first argument to declaring the component as a shorthand (e.g. tray.button("Click me") instead of tray.button({label: "Click me"}))
}

func defineComponent(vm *goja.Runtime, call goja.FunctionCall, t string, propDefs []ComponentProp) goja.Value {
	component := Component{
		ID:    uuid.New().String(),
		Type:  t,
		Props: make(map[string]interface{}),
	}

	propsList := make(map[string]interface{})
	propDefsMap := make(map[string]*ComponentProp)

	var shorthandProp *ComponentProp
	for _, propDef := range propDefs {

		propDefsMap[propDef.Name] = &propDef

		if propDef.OptionalFirstArg {
			shorthandProp = &propDef
		}
	}

	if len(call.Arguments) > 0 {
		// Check if the first argument is the type of the shorthand
		hasShorthand := false
		if shorthandProp != nil {
			switch shorthandProp.Type {
			case "string":
				if _, ok := call.Argument(0).Export().(string); ok {
					propsList[shorthandProp.Name] = call.Argument(0).Export().(string)
					hasShorthand = true
				}
			case "boolean":
				if _, ok := call.Argument(0).Export().(bool); ok {
					propsList[shorthandProp.Name] = call.Argument(0).Export().(bool)
					hasShorthand = true
				}
			case "array":
				if _, ok := call.Argument(0).Export().([]interface{}); ok {
					propsList[shorthandProp.Name] = call.Argument(0).Export().([]interface{})
					hasShorthand = true
				}
			}
			if hasShorthand {
				// Get the rest of the props from the second argument
				if len(call.Arguments) > 1 {
					rest, ok := call.Argument(1).Export().(map[string]interface{})
					if ok {
						// Only add props that are defined in the propDefs
						for k, v := range rest {
							if _, ok := propDefsMap[k]; ok {
								propsList[k] = v
							}
						}
					}
				}
			}
		}

		if !hasShorthand {
			propsArg, ok := call.Argument(0).Export().(map[string]interface{})
			if ok {
				for k, v := range propsArg {
					if _, ok := propDefsMap[k]; ok {
						propsList[k] = v
					} else {
						// util.SpewMany(k, fmt.Sprintf("%T", v))
					}
				}
			}
		}
	}

	// Validate props
	for _, propDef := range propDefs {
		// If a prop is required and no value is provided, panic
		if propDef.Required && len(propsList) == 0 {
			panic(vm.NewTypeError(fmt.Sprintf("%s is required", propDef.Name)))
		}

		// Validate the prop if the prop is defined
		if propDef.Validate != nil {
			if val, ok := propsList[propDef.Name]; ok {
				err := propDef.Validate(val)
				if err != nil {
					panic(vm.NewTypeError(err.Error()))
				}
			}
		}

		// Set a default value if the prop is not provided
		if _, ok := propsList[propDef.Name]; !ok && propDef.Default != nil {
			propsList[propDef.Name] = propDef.Default
		}
	}

	// Set the props
	for k, v := range propsList {
		component.Props[k] = v
	}

	return vm.ToValue(component)
}

// Helper function to create a validation function for a specific type
func validateType(expectedType string) func(interface{}) error {
	return func(value interface{}) error {
		switch expectedType {
		case "string":
			_, ok := value.(string)
			if !ok {
				if value == nil {
					return nil
				}
				return fmt.Errorf("expected string, got %T", value)
			}
			return nil
		case "number":
			_, ok := value.(float64)
			if !ok {
				_, ok := value.(int64)
				if !ok {
					if value == nil {
						return nil
					}
					return fmt.Errorf("expected number, got %T", value)
				}
				return nil
			}
			return nil
		case "boolean":
			_, ok := value.(bool)
			if !ok {
				if value == nil {
					return nil
				}
				return fmt.Errorf("expected boolean, got %T", value)
			}
			return nil
		case "array":
			_, ok := value.([]interface{})
			if !ok {
				if value == nil {
					return nil
				}
				return fmt.Errorf("expected array, got %T", value)
			}
			return nil
		case "object":
			_, ok := value.(map[string]interface{})
			if !ok {
				if value == nil {
					return nil
				}
				return fmt.Errorf("expected object, got %T", value)
			}
			return nil
		default:
			return fmt.Errorf("invalid type: %s", expectedType)
		}
	}
}

// componentDiff compares two component trees and returns a new component tree that preserves the ID of old components that did not change.
// It also recursively handles props and items arrays.
//
// This is important to preserve state between renders in React.
func (c *ComponentManager) componentDiff(old, new interface{}) (ret interface{}) {
	defer func() {
		if r := recover(); r != nil {
			// If a panic occurs, return the new component tree
			ret = new
		}
	}()

	if old == nil || new == nil {
		return new
	}

	// Handle maps (components)
	if oldMap, ok := old.(map[string]interface{}); ok {
		if newMap, ok := new.(map[string]interface{}); ok {
			// If types match and it's a component (has "type" field), preserve ID
			if oldType, hasOldType := oldMap["type"]; hasOldType {
				if newType, hasNewType := newMap["type"]; hasNewType && oldType == newType {
					// Preserve the ID from the old component
					if oldID, hasOldID := oldMap["id"]; hasOldID {
						newMap["id"] = oldID
					}

					// Recursively handle props
					if oldProps, hasOldProps := oldMap["props"].(map[string]interface{}); hasOldProps {
						if newProps, hasNewProps := newMap["props"].(map[string]interface{}); hasNewProps {
							// Special handling for items array in props
							if oldItems, ok := oldProps["items"].([]interface{}); ok {
								if newItems, ok := newProps["items"].([]interface{}); ok {
									newProps["items"] = c.componentDiff(oldItems, newItems)
								}
							}
							// Handle other props
							for k, v := range newProps {
								if k != "items" { // Skip items as we already handled it
									if oldV, exists := oldProps[k]; exists {
										newProps[k] = c.componentDiff(oldV, v)
									}
								}
							}
							newMap["props"] = newProps
						}
					}
				}
			}
			return newMap
		}
	}

	// Handle arrays
	if oldArr, ok := old.([]interface{}); ok {
		if newArr, ok := new.([]interface{}); ok {
			// Create a new array to store the diffed components
			result := make([]interface{}, len(newArr))

			// First, try to match components by key if available
			oldKeyMap := make(map[string]interface{})
			for _, oldComp := range oldArr {
				if oldMap, ok := oldComp.(map[string]interface{}); ok {
					if key, ok := oldMap["key"].(string); ok && key != "" {
						oldKeyMap[key] = oldComp
					}
				}
			}

			// Process each new component
			for i, newComp := range newArr {
				matched := false

				// Try to match by key first
				if newMap, ok := newComp.(map[string]interface{}); ok {
					if key, ok := newMap["key"].(string); ok && key != "" {
						if oldComp, exists := oldKeyMap[key]; exists {
							// Found a match by key
							result[i] = c.componentDiff(oldComp, newComp)
							matched = true
							// t.ctx.logger.Debug().
							// 	Str("key", key).
							// 	Str("type", fmt.Sprintf("%v", newMap["type"])).
							// 	Msg("Component matched by key")
						}
					}
				}

				// If no key match, try to match by position and type
				if !matched && i < len(oldArr) {
					oldComp := oldArr[i]
					oldType, newType := "", ""

					if oldMap, ok := oldComp.(map[string]interface{}); ok {
						if t, ok := oldMap["type"].(string); ok {
							oldType = t
						}
					}
					if newMap, ok := newComp.(map[string]interface{}); ok {
						if t, ok := newMap["type"].(string); ok {
							newType = t
						}
					}

					if oldType != "" && oldType == newType {
						result[i] = c.componentDiff(oldComp, newComp)
						matched = true
						// t.ctx.logger.Debug().
						// 	Str("type", oldType).
						// 	Msg("Component matched by type and position")
					}
				}

				// If no match found, use the new component as is
				if !matched {
					result[i] = newComp
					// if newMap, ok := newComp.(map[string]interface{}); ok {
					// 	t.ctx.logger.Debug().
					// 		Str("type", fmt.Sprintf("%v", newMap["type"])).
					// 		Msg("New component added")
					// }
				}
			}

			// Log removed components
			// if len(oldArr) > len(newArr) {
			// 	for i := len(newArr); i < len(oldArr); i++ {
			// 		if oldMap, ok := oldArr[i].(map[string]interface{}); ok {
			// 			t.ctx.logger.Debug().
			// 				Str("type", fmt.Sprintf("%v", oldMap["type"])).
			// 				Msg("Component removed")
			// 		}
			// 	}
			// }

			return result
		}
	}

	return new
}
