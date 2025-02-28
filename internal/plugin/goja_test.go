package plugin

import (
	"errors"
	"seanime/internal/util"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
)

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func TestDivideFunction(t *testing.T) {
	vm := goja.New()
	vm.Set("divide", divide)

	// Case 1: Successful division
	result, err := vm.RunString("divide(10, 3);")
	assert.NoError(t, err)
	assert.Equal(t, 3.3333333333333335, result.Export())

	// Case 2: Division by zero should throw an exception
	_, err = vm.RunString("divide(10, 0);")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "division by zero")

	// Case 3: Handling error with try-catch in JS
	result, err = vm.RunString(`
		try {
			divide(10, 0);
		} catch (e) {
			e.toString();
		}
	`)
	assert.NoError(t, err)
	assert.Equal(t, "GoError: division by zero", result.Export())
}

func multipleReturns() (int, string, float64) {
	return 42, "hello", 3.14
}

func TestMultipleReturns(t *testing.T) {
	vm := goja.New()
	vm.Set("multiReturn", multipleReturns)

	v, err := vm.RunString("multiReturn();")
	assert.NoError(t, err)
	util.Spew(v.Export())
}
