package extension_repo

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"seanime/internal/extension"
	"time"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type gojaExtensionImpl struct {
	ext      *extension.Extension
	vm       *goja.Runtime
	logger   *zerolog.Logger
	classObj *goja.Object
}

func (g *gojaExtensionImpl) error(err error, msg ...string) error {
	if len(msg) > 0 {
		g.logger.Error().Err(err).Str("id", g.ext.ID).Msgf("extensions: %s, %v", msg[0], err)
		return fmt.Errorf("%s, %v", msg[0], err)
	}
	g.logger.Error().Err(err).Str("id", g.ext.ID).Msgf("extensions: Unexpected error, %v", err)
	return err
}

// getClassMethod returns the classObj method by name
func (g *gojaExtensionImpl) getClassMethod(name string) (goja.Callable, error) {
	method, ok := goja.AssertFunction(g.classObj.Get(name))
	if !ok {
		return nil, g.error(fmt.Errorf("failed to get '%s' function", name))
	}
	return method, nil
}

// callClassMethod calls the classObj method by name with the provided arguments
func (g *gojaExtensionImpl) callClassMethod(name string, args ...goja.Value) (ret goja.Value, err error) {
	method, err := g.getClassMethod(name)
	if err != nil {
		return nil, err
	}

	value, err := method(g.classObj, args...)
	if err != nil {
		return nil, g.error(err, fmt.Sprintf("failed to call '%s' function", name))
	}

	return value, nil
}

func (g *gojaExtensionImpl) waitForPromise(value goja.Value) (goja.Value, error) {
	promise, ok := value.Export().(*goja.Promise)
	if !ok {
		return nil, g.error(fmt.Errorf("value is not a promise"))
	}

	doneCh := make(chan struct{})

	go func() {
		for promise.State() == goja.PromiseStatePending {
			time.Sleep(10 * time.Millisecond)
		}
		close(doneCh)
	}()

	<-doneCh

	if promise.State() == goja.PromiseStateRejected {
		err := promise.Result()
		return nil, g.error(fmt.Errorf("%v", err), "promise rejected")
	}

	res := promise.Result()

	if res == nil || goja.IsUndefined(res) {
		return nil, g.error(fmt.Errorf("promise result is undefined"))
	}

	return res, nil
}

func (g *gojaExtensionImpl) unmarshalValue(value goja.Value, ret interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = g.error(fmt.Errorf("%v", r), "failed to unmarshal result")
		}
	}()

	jsonData, err := json.Marshal(value.Export())
	if err != nil {
		return g.error(err, "failed to marshal result")
	}

	err = json.Unmarshal(jsonData, &ret)
	if err != nil {
		return g.error(err, fmt.Sprintf("failed to unmarshal result: %s", string(jsonData)))
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// structToMap converts a struct to "JSON-like" map for Goja extensions
func structToMap(obj interface{}) map[string]interface{} {
	// Convert the struct to a map
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil
	}

	var data map[string]interface{}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil
	}

	return data
}
