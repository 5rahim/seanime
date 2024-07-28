package extension_repo

import (
	"context"
	"github.com/traefik/yaegi/interp"
	"reflect"
	"time"
)

const (
	MsgYaegiFailedToEvaluateExtensionCode = "extensions: Failed to evaluate extension source code"
	MsgYaegiFailedToInstantiateExtension  = "extensions: Failed to instantiate extension, the extension is incompatible with the expected interface"
)

func yaegiEval(i *interp.Interpreter, src string) (reflect.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return i.EvalWithContext(ctx, src)
}
