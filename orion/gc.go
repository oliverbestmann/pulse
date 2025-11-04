package orion

import (
	"log/slog"
	"reflect"
	"runtime"
)

type releaser interface{ Release() }

// RegisterWithGC automatically calls Release on value if
// the value is garbage collected
func RegisterWithGC[T releaser](value T) T {
	if runtime.GOOS == "js" {
		// js values are garbage collected anyways, no need to
		// register the Finalizer
		return value
	}

	runtime.SetFinalizer(value, releaseNow[T])

	return value
}

func releaseNow[T releaser](value T) {
	typ := reflect.TypeOf(value).String()
	slog.Debug("Releasing garbage collected instance", slog.String("type", typ))

	value.Release()
}
