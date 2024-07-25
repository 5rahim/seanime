package extension_repo

import (
	"context"
	"reflect"
	"time"
)

const (
	MsgYaegiFailedToEvaluateExtensionCode = "extensions: Failed to evaluate extension source code"
	MsgYaegiFailedToInstantiateExtension  = "extensions: Failed to instantiate extension, the extension is incompatible with the expected interface"
)

func (r *Repository) yaegiEval(src string) (reflect.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return r.yaegiInterp.EvalWithContext(ctx, src)
}
