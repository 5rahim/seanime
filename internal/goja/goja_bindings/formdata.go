package goja_bindings

import (
	"bytes"
	"io"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/dop251/goja"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// formData
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func BindFormData(vm *goja.Runtime) error {
	err := vm.Set("FormData", func(call goja.ConstructorCall) *goja.Object {
		fd := newFormData(vm)

		instanceValue := vm.ToValue(fd).(*goja.Object)
		instanceValue.SetPrototype(call.This.Prototype())

		return instanceValue
	})
	if err != nil {
		return err
	}
	return nil
}

type formData struct {
	runtime    *goja.Runtime
	buf        *bytes.Buffer
	writer     *multipart.Writer
	fieldNames map[string]struct{}
	values     map[string][]string
	closed     bool
}

func newFormData(runtime *goja.Runtime) *formData {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	return &formData{
		runtime:    runtime,
		buf:        buf,
		writer:     writer,
		fieldNames: make(map[string]struct{}),
		values:     make(map[string][]string),
		closed:     false,
	}
}

func (fd *formData) Append(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot append to closed FormData")
	}

	fieldName := call.Argument(0).String()
	value := call.Argument(1).String()

	fieldName = strings.TrimSpace(fieldName)
	fd.values[fieldName] = append(fd.values[fieldName], value)

	if _, exists := fd.fieldNames[fieldName]; !exists {
		fd.fieldNames[fieldName] = struct{}{}
		writer, err := fd.writer.CreateFormField(fieldName)
		if err != nil {
			return fd.runtime.ToValue(err.Error())
		}
		_, err = writer.Write([]byte(value))
		if err != nil {
			return fd.runtime.ToValue(err.Error())
		}
	}

	return goja.Undefined()
}

func (fd *formData) Delete(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot delete from closed FormData")
	}

	fieldName := call.Argument(0).String()
	fieldName = strings.TrimSpace(fieldName)

	delete(fd.fieldNames, fieldName)
	delete(fd.values, fieldName)

	return goja.Undefined()
}

func (fd *formData) Entries(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot get entries from closed FormData")
	}

	iter := fd.runtime.NewArray()
	index := 0
	for key, values := range fd.values {
		for _, value := range values {
			entry := fd.runtime.NewObject()
			entry.Set("0", key)
			entry.Set("1", value)
			iter.Set(strconv.Itoa(index), entry)
			index++
		}
	}

	return iter
}

func (fd *formData) Get(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot get value from closed FormData")
	}

	fieldName := call.Argument(0).String()
	fieldName = strings.TrimSpace(fieldName)

	if values, exists := fd.values[fieldName]; exists && len(values) > 0 {
		return fd.runtime.ToValue(values[0])
	}

	return goja.Undefined()
}

func (fd *formData) GetAll(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot get all values from closed FormData")
	}

	fieldName := call.Argument(0).String()
	fieldName = strings.TrimSpace(fieldName)

	iter := fd.runtime.NewArray()
	if values, exists := fd.values[fieldName]; exists {
		for i, value := range values {
			iter.Set(strconv.Itoa(i), value)
		}
	}

	return iter
}

func (fd *formData) Has(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot check key in closed FormData")
	}

	fieldName := call.Argument(0).String()
	fieldName = strings.TrimSpace(fieldName)

	_, exists := fd.fieldNames[fieldName]
	return fd.runtime.ToValue(exists)
}

func (fd *formData) Keys(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot get keys from closed FormData")
	}

	iter := fd.runtime.NewArray()
	index := 0
	for key := range fd.fieldNames {
		iter.Set(strconv.Itoa(index), key)
		index++
	}

	return iter
}

func (fd *formData) Set(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot set value in closed FormData")
	}

	fieldName := call.Argument(0).String()
	value := call.Argument(1).String()

	fieldName = strings.TrimSpace(fieldName)
	fd.values[fieldName] = []string{value}

	if _, exists := fd.fieldNames[fieldName]; !exists {
		fd.fieldNames[fieldName] = struct{}{}
		writer, err := fd.writer.CreateFormField(fieldName)
		if err != nil {
			return fd.runtime.ToValue(err.Error())
		}
		_, err = writer.Write([]byte(value))
		if err != nil {
			return fd.runtime.ToValue(err.Error())
		}
	}

	return goja.Undefined()
}

func (fd *formData) Values(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot get values from closed FormData")
	}

	iter := fd.runtime.NewArray()
	index := 0
	for _, values := range fd.values {
		for _, value := range values {
			iter.Set(strconv.Itoa(index), value)
			index++
		}
	}

	return iter
}

func (fd *formData) GetContentType() goja.Value {
	if !fd.closed {
		fd.writer.Close()
		fd.closed = true
	}
	return fd.runtime.ToValue(fd.writer.FormDataContentType())
}

func (fd *formData) GetBuffer() (io.Reader, *multipart.Writer) {
	if !fd.closed {
		fd.writer.Close()
		fd.closed = true
	}
	return bytes.NewReader(fd.buf.Bytes()), fd.writer
}
