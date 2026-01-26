package litesqltest

import (
	"testing"

	"github.com/shoenig/test/must"
)

func Test_Open(t *testing.T) {
	t.Parallel()

	ldb := Open(t)

	m, perr := ldb.Pragmas(t.Context())
	must.NoError(t, perr)
	must.MapLen(t, 8, m)
	must.Eq(t, "UTF-8", m["encoding"])
	must.Eq(t, "500", m["busy_timeout"])
	must.Eq(t, "1", m["foreign_keys"])
	must.Eq(t, "off", m["journal_mode"])
	must.Eq(t, "-4000", m["cache_size"])
	must.Eq(t, "2", m["auto_vacuum"])
	must.Eq(t, "1", m["synchronous"])
	must.Eq(t, "4096", m["page_size"])
}
