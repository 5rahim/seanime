package plugin_ui

import (
	"github.com/dop251/goja"
	"github.com/google/uuid"
)

type FormManager struct {
	ctx *Context
}

func NewFormManager(ctx *Context) *FormManager {
	return &FormManager{
		ctx: ctx,
	}
}

type FormField struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Label       string                 `json:"label"`
	Placeholder string                 `json:"placeholder,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Options     []FormFieldOption      `json:"options,omitempty"`
	Props       map[string]interface{} `json:"props,omitempty"`
}

type FormFieldOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

type Form struct {
	Name    string    `json:"name"`
	ID      string    `json:"id"`
	Type    string    `json:"type"`
	Props   FormProps `json:"props"`
	manager *FormManager
}

type FormProps struct {
	Name   string      `json:"name"`
	Fields []FormField `json:"fields"`
}

// jsNewForm
//
//	Example:
//	const form = tray.newForm("form-1")
func (f *FormManager) jsNewForm(call goja.FunctionCall) goja.Value {
	name, ok := call.Argument(0).Export().(string)
	if !ok {
		f.ctx.handleTypeError("newForm requires a name")
	}

	form := &Form{
		Name:    name,
		ID:      uuid.New().String(),
		Type:    "form",
		Props:   FormProps{Fields: make([]FormField, 0), Name: name},
		manager: f,
	}

	formObj := f.ctx.vm.NewObject()

	// Form methods
	formObj.Set("render", form.jsRender)
	formObj.Set("onSubmit", form.jsOnSubmit)

	// Field creation methods
	formObj.Set("inputField", form.jsInputField)
	formObj.Set("numberField", form.jsNumberField)
	formObj.Set("selectField", form.jsSelectField)
	formObj.Set("checkboxField", form.jsCheckboxField)
	formObj.Set("radioField", form.jsRadioField)
	formObj.Set("dateField", form.jsDateField)
	formObj.Set("switchField", form.jsSwitchField)
	formObj.Set("submitButton", form.jsSubmitButton)
	formObj.Set("reset", form.jsReset)
	formObj.Set("setValues", form.jsSetValues)

	return formObj
}

func (f *Form) jsRender(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		f.manager.ctx.handleTypeError("render requires a config object")
	}

	config, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		f.manager.ctx.handleTypeError("render requires a config object")
	}

	if fields, ok := config["fields"].([]interface{}); ok {
		f.Props.Fields = make([]FormField, 0)
		for _, field := range fields {
			if fieldMap, ok := field.(FormField); ok {
				f.Props.Fields = append(f.Props.Fields, fieldMap)
			}
		}
	}

	return f.manager.ctx.vm.ToValue(f)
}

func (f *Form) jsOnSubmit(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		f.manager.ctx.handleTypeError("onSubmit requires a callback function")
	}

	callback, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		f.manager.ctx.handleTypeError("onSubmit requires a callback function")
	}

	eventListener := f.manager.ctx.RegisterEventListener(ClientFormSubmittedEvent)

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		var payload ClientFormSubmittedEventPayload
		if event.ParsePayloadAs(ClientFormSubmittedEvent, &payload) && payload.FormName == f.Name {
			f.manager.ctx.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), f.manager.ctx.vm.ToValue(payload.Data))
				return err
			})
		}
	})

	// go func() {
	// 	for event := range eventListener.Channel {
	// 		if event.ParsePayloadAs(ClientFormSubmittedEvent, &payload) {
	// 			if payload.FormName == f.Name {
	// 				f.manager.ctx.scheduler.ScheduleAsync(func() error {
	// 					_, err := callback(goja.Undefined(), f.manager.ctx.vm.ToValue(payload.Data))
	// 					if err != nil {
	// 						f.manager.ctx.logger.Error().Err(err).Msg("error running form submit callback")
	// 					}
	// 					return err
	// 				})
	// 			}
	// 		}
	// 	}
	// }()

	return goja.Undefined()
}

func (f *Form) jsReset(call goja.FunctionCall) goja.Value {
	fieldToReset := ""
	if len(call.Arguments) > 0 {
		var ok bool
		fieldToReset, ok = call.Argument(0).Export().(string)
		if !ok {
			f.manager.ctx.handleTypeError("reset requires a field name")
		}
	}

	f.manager.ctx.SendEventToClient(ServerFormResetEvent, ServerFormResetEventPayload{
		FormName:     f.Name,
		FieldToReset: fieldToReset,
	})

	return goja.Undefined()
}

func (f *Form) jsSetValues(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		f.manager.ctx.handleTypeError("setValues requires a config object")
	}

	props, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		f.manager.ctx.handleTypeError("setValues requires a config object")
	}

	f.manager.ctx.SendEventToClient(ServerFormSetValuesEvent, ServerFormSetValuesEventPayload{
		FormName: f.Name,
		Data:     props,
	})

	return goja.Undefined()
}

func (f *Form) createField(fieldType string, props map[string]interface{}) goja.Value {
	nameRaw, ok := props["name"]
	name := ""
	if ok {
		name, ok = nameRaw.(string)
		if !ok {
			f.manager.ctx.handleTypeError("name must be a string")
		}
	}
	label := ""
	labelRaw, ok := props["label"]
	if ok {
		label, ok = labelRaw.(string)
		if !ok {
			f.manager.ctx.handleTypeError("label must be a string")
		}
	}
	placeholder, ok := props["placeholder"]
	if ok {
		placeholder, ok = placeholder.(string)
		if !ok {
			f.manager.ctx.handleTypeError("placeholder must be a string")
		}
	}
	field := FormField{
		ID:      uuid.New().String(),
		Type:    fieldType,
		Name:    name,
		Label:   label,
		Value:   props["value"],
		Options: nil,
	}

	// Handle options if present
	if options, ok := props["options"].([]interface{}); ok {
		fieldOptions := make([]FormFieldOption, len(options))
		for i, opt := range options {
			if optMap, ok := opt.(map[string]interface{}); ok {
				fieldOptions[i] = FormFieldOption{
					Label: optMap["label"].(string),
					Value: optMap["value"],
				}
			}
		}
		field.Options = fieldOptions
	}

	return f.manager.ctx.vm.ToValue(field)
}

func (f *Form) jsInputField(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		f.manager.ctx.handleTypeError("inputField requires a config object")
	}

	props, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		f.manager.ctx.handleTypeError("inputField requires a config object")
	}

	return f.createField("input", props)
}

func (f *Form) jsNumberField(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		f.manager.ctx.handleTypeError("numberField requires a config object")
	}

	props, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		f.manager.ctx.handleTypeError("numberField requires a config object")
	}

	return f.createField("number", props)
}

func (f *Form) jsSelectField(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		f.manager.ctx.handleTypeError("selectField requires a config object")
	}

	props, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		f.manager.ctx.handleTypeError("selectField requires a config object")
	}

	return f.createField("select", props)
}

func (f *Form) jsCheckboxField(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		f.manager.ctx.handleTypeError("checkboxField requires a config object")
	}

	props, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		f.manager.ctx.handleTypeError("checkboxField requires a config object")
	}

	return f.createField("checkbox", props)
}

func (f *Form) jsSwitchField(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		f.manager.ctx.handleTypeError("switchField requires a config object")
	}

	props, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		f.manager.ctx.handleTypeError("switchField requires a config object")
	}

	return f.createField("switch", props)
}

func (f *Form) jsRadioField(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		f.manager.ctx.handleTypeError("radioField requires a config object")
	}

	props, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		f.manager.ctx.handleTypeError("radioField requires a config object")
	}

	return f.createField("radio", props)
}

func (f *Form) jsDateField(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		f.manager.ctx.handleTypeError("dateField requires a config object")
	}

	props, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		f.manager.ctx.handleTypeError("dateField requires a config object")
	}

	return f.createField("date", props)
}

func (f *Form) jsSubmitButton(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		f.manager.ctx.handleTypeError("submitButton requires a config object")
	}

	props, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		f.manager.ctx.handleTypeError("submitButton requires a config object")
	}

	return f.createField("submit", props)
}
