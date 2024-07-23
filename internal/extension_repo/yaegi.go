package extension_repo

import (
	"context"
	"reflect"
	"time"
)

const (
	MsgYaegiFailedToEvaluateExtensionCode = "extension repo: Failed to evaluate extension source code"
	MsgYaegiFailedToInstantiateExtension  = "extension repo: Failed to instantiate extension, the extension is incompatible with the expected interface"
)

func (r *Repository) yaegiEval(src string) (reflect.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return r.yaegiInterp.EvalWithContext(ctx, src)
}
