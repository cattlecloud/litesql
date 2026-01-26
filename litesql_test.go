package litesql

import (
	"testing"

	"cattlecloud.net/go/scope"
	"github.com/shoenig/test/must"
)

func TestLiteDB_Pragmas(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), timeout)
	defer cancel()

	ldb := testSimple(t)

	m, perr := ldb.Pragmas(ctx)
	must.NoError(t, perr)

	must.Eq(t, "UTF-8", m["encoding"])
	must.Eq(t, "1000", m["busy_timeout"])
	must.Eq(t, "1", m["foreign_keys"])
	must.Eq(t, "off", m["journal_mode"])
	must.Eq(t, "-4000", m["cache_size"])
	must.Eq(t, "2", m["auto_vacuum"])
	must.Eq(t, "1", m["synchronous"])
	must.Eq(t, "4096", m["page_size"])
}
