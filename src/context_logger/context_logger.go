package context_logger

import (
	"path"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

var (
	// G is an alias for GetLogger
	G = GetLogger

	// L is an alias for the standard logger
	L = logrus.NewEntry(logrus.StandardLogger())
)

type (
	loggerKey struct{}
	moduleKey struct{}
)

// returns a new context with the provided logger
// Use in combination with logger.WithField(s) for great effect
func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// returns the current logger from the context
// If logger is available, the default logger is returned
func GetLogger(ctx context.Context) *logrus.Entry {
	logger := ctx.Value(loggerKey{})

	if logger == nil {
		return L
	}

	return logger.(*logrus.Entry)
}

// adds the module to the context, appending it with a slash if a module already
// exists. A moudle is just an roughly correlated defined by the call tree for
// given context
//
// Modules represent the call path. If the new module and last modlue and last modlue
// are the same, a new modlue entry will noe be created. If the new modlue and old older
// modlue are the same bur separated bu other modlues, the cycle will be represented the module path
func WithModlue(ctx context.Context, module string) context.Context {
	parent := GetModulePath(ctx)

	if parent != "" {
		// don't re-append module when module is the same
		if path.Base(parent) == module {
			return ctx
		}

		module = path.Join(parent, module)
	}

	ctx = WithLogger(ctx, GetLogger(ctx).WithField("module", module))
	return context.WithValue(ctx, moduleKey{}, module)
}

// returns the module path for the provided context. If no module is set
// am empty string is returned
func GetModulePath(ctx context.Context) string {
	module := ctx.Value(moduleKey{})
	if module == nil {
		return ""
	}

	return module.(string)
}
