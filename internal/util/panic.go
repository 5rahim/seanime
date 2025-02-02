package util

import (
	"errors"
	"github.com/rs/zerolog/log"
	"runtime/debug"
	"sync"
)

var printLock = sync.Mutex{}

func printRuntimeError(r any, module string) string {
	printLock.Lock()
	debugStr := string(debug.Stack())
	logger := NewLogger()
	log.Error().Msgf("go: PANIC RECOVERY")
	if module != "" {
		log.Error().Msgf("go: Runtime error in \"%s\"", module)
	}
	log.Error().Msgf("go: A runtime error occurred, please send the logs to the developer\n")
	log.Printf("go: ========================================= Stack Trace =========================================\n")
	logger.Error().Msgf("%+v\n\n%+v", r, debugStr)
	log.Printf("go: ===================================== End of Stack Trace ======================================\n")
	printLock.Unlock()
	return debugStr
}

func HandlePanicWithError(err *error) {
	if r := recover(); r != nil {
		*err = errors.New("fatal error occurred, please report this issue")
		printRuntimeError(r, "")
	}
}

func HandlePanicInModuleWithError(module string, err *error) {
	if r := recover(); r != nil {
		*err = errors.New("fatal error occurred, please report this issue")
		printRuntimeError(r, module)
	}
}

func HandlePanicThen(f func()) {
	if r := recover(); r != nil {
		f()
		printRuntimeError(r, "")
	}
}

func HandlePanicInModuleThen(module string, f func()) {
	if r := recover(); r != nil {
		f()
		printRuntimeError(r, module)
	}
}

func HandlePanicInModuleThenS(module string, f func(stackTrace string)) {
	if r := recover(); r != nil {
		str := printRuntimeError(r, module)
		f(str)
	}
}

func Recover() {
	if r := recover(); r != nil {
		printRuntimeError(r, "")
	}
}

func RecoverInModule(module string) {
	if r := recover(); r != nil {
		printRuntimeError(r, module)
	}
}
