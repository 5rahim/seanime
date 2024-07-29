package extension_repo

import (
	"context"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"reflect"
	"seanime/internal/yaegi_interp"
	"time"
)

const (
	MsgYaegiFailedToEvaluateExtensionCode = "extensions: Failed to evaluate extension source code"
	MsgYaegiFailedToInstantiateExtension  = "extensions: Failed to instantiate extension, the extension is incompatible with the expected interface"
)

func (r *Repository) loadYaegiInterpreter() {
	i := interp.New(interp.Options{
		Unrestricted: false,
	})

	symbols := stdlib.Symbols
	// Remove symbols from stdlib that are risky to give to extensions
	delete(symbols, "os/os")
	delete(symbols, "io/fs/fs")
	delete(symbols, "os/exec/exec")
	delete(symbols, "os/signal/signal")
	delete(symbols, "os/user/user")
	delete(symbols, "os/signal/signal")
	delete(symbols, "io/ioutil/ioutil")
	delete(symbols, "runtime/runtime")
	delete(symbols, "syscall/syscall")
	delete(symbols, "archive/tar/tar")
	delete(symbols, "archive/zip/zip")
	delete(symbols, "compress/gzip/gzip")
	delete(symbols, "compress/zlib/zlib")

	if err := i.Use(symbols); err != nil {
		r.logger.Fatal().Err(err).Msg("extensions: Failed to load yaegi stdlib")
	}

	// Load the extension symbols
	err := i.Use(yaegi_interp.Symbols)
	if err != nil {
		r.logger.Fatal().Err(err).Msg("extensions: Failed to load extension symbols")
	}

	r.yaegiInterp = i
}

func yaegiEval(i *interp.Interpreter, src string) (reflect.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return i.EvalWithContext(ctx, src)
}
