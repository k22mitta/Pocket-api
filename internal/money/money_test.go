package money

import "testing"

func TestRound2(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{250.00 - 240.82, 9.18}, // raw float64 subtraction yields 9.180000000000007
		{9.180000000000007, 9.18},
		{-9.180000000000007, -9.18},
		{0, 0},
		{100.005, 100.01}, // half-cent rounds away from zero
		{-100.005, -100.01},
		{1234.5, 1234.5},
	}
	for _, c := range cases {
		got := Round2(c.in)
		if got != c.want {
			t.Errorf("Round2(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

// TestRound2NoDriftAcrossManyAdditions guards against accumulated float64
// error in loops that add many small amounts (e.g. summing a month of
// transactions), which is exactly how the budgets "spent" figure is built.
func TestRound2NoDriftAcrossManyAdditions(t *testing.T) {
	sum := 0.0
	for i := 0; i < 1000; i++ {
		sum += 0.01
	}
	got := Round2(sum)
	if got != 10.00 {
		t.Errorf("Round2(sum of 1000x 0.01) = %v, want 10", got)
	}
}
