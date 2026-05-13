package litesql

import (
	"testing"
	"time"

	"cattlecloud.net/go/scope"
	"github.com/shoenig/test/must"
)

func Test_Optimize(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), 10*time.Second)
	defer cancel()

	ldb := testSimple(t)
	err := ldb.Optimize(ctx)
	must.NoError(t, err)
}

func Test_Analyze(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), 10*time.Second)
	defer cancel()

	ldb := testSimple(t)
	err := ldb.Analyze(ctx)
	must.NoError(t, err)
}

func Test_Vacuum(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), 10*time.Second)
	defer cancel()

	ldb := testSimple(t)
	err := ldb.Vacuum(ctx)
	must.NoError(t, err)
}
