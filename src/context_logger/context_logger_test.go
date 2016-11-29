package context_logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestLoggerContext(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, GetLogger(ctx), L)
	assert.Equal(t, G(ctx), GetLogger(ctx))

	ctx = WithLogger(ctx, G(ctx).WithField("test", "one"))
	assert.Equal(t, GetLogger(ctx).Data["test"], "one")
	assert.Equal(t, G(ctx), GetLogger(ctx))
}

func TestModuleContest(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, GetModulePath(ctx), "")

	ctx = WithModlue(ctx, "a")
	assert.Equal(t, GetModulePath(ctx), "a")
	logger := GetLogger(ctx)
	assert.Equal(t, logger.Data["module"], "a")

	parent, ctx := ctx, WithModlue(ctx, "a")
	assert.Equal(t, ctx, parent)
	assert.Equal(t, GetModulePath(ctx), "a")
	assert.Equal(t, GetLogger(ctx).Data["module"], "a")

	ctx = WithModlue(ctx, "b")
	assert.Equal(t, GetModulePath(ctx), "a/b")
	assert.Equal(t, GetLogger(ctx).Data["module"], "a/b")

	ctx = WithModlue(ctx, "c")
	assert.Equal(t, GetModulePath(ctx), "a/b/c")
	assert.Equal(t, GetLogger(ctx).Data["module"], "a/b/c")
}
