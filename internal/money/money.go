// Package money holds the one rounding helper every dollar amount must pass
// through before it's returned from the API or fed into an aggregate SQL
// query result. Go float64 arithmetic (e.g. amount_limit - spent) can leave
// values like 9.180000000000007 instead of 9.18; Postgres NUMERIC columns are
// exact, but that exactness is lost the moment a value is scanned into a Go
// float64 and combined with another float64. Rounding at the boundary is a
// pragmatic fix short of migrating the whole schema to integer cents.
package money

import "math"

// Round2 rounds a float64 to 2 decimal places using round-half-away-from-zero,
// matching how currency amounts are conventionally rounded.
func Round2(v float64) float64 {
	if v >= 0 {
		return math.Round(v*100) / 100
	}
	return -(math.Round(-v*100) / 100)
}
