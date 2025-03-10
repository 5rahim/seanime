package goja_bindings

import (
	"seanime/internal/util"
	"testing"

	"github.com/dop251/goja"
	gojabuffer "github.com/dop251/goja_nodejs/buffer"
	gojarequire "github.com/dop251/goja_nodejs/require"
	"github.com/stretchr/testify/require"
)

func TestGojaFormData(t *testing.T) {
	vm := goja.New()
	defer vm.ClearInterrupt()

	BindFormData(vm)

	registry := new(gojarequire.Registry)
	registry.Enable(vm)
	gojabuffer.Enable(vm)
	BindConsole(vm, util.NewLogger())

	_, err := vm.RunString(`
var fd = new FormData();
fd.append("name", "John Doe");
fd.append("age", 30);

console.log("Has 'name':", fd.has("name")); // true
console.log("Get 'name':", fd.get("name")); // John Doe
console.log("GetAll 'name':", fd.getAll("name")); // ["John Doe"]
console.log("Keys:", Array.from(fd.keys())); // ["name", "age"]
console.log("Values:", Array.from(fd.values())); // ["John Doe", 30]

fd.delete("name");
console.log("Has 'name' after delete:", fd.has("name")); // false

console.log("Entries:");
for (let entry of fd.entries()) {
	console.log(entry[0], entry[1]);
}

var contentType = fd.getContentType();
var buffer = fd.getBuffer();
console.log("Content-Type:", contentType);
console.log("Buffer:", buffer);
	`)
	require.NoError(t, err)
}
