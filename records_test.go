package litesql

import (
	"testing"

	"github.com/shoenig/test/must"
)

func TestID_String(t *testing.T) {
	t.Parallel()

	cases := []struct {
		value ID
		exp   string
	}{
		{
			value: 1,
			exp:   "id:1",
		},
		{
			value: ExecFailure,
			exp:   "exec-failure",
		},
		{
			value: TxFailure,
			exp:   "tx-failure",
		},
	}

	for _, tc := range cases {
		result := tc.value.String()
		must.Eq(t, tc.exp, result)
	}
}
