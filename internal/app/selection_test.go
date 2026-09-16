package app

import "testing"

func TestSelectionGridNavigation(t *testing.T) {
	cases := []struct{ from, dx, dy, want int }{{0, -1, 0, 4}, {4, 1, 0, 0}, {0, 0, 1, 5}, {7, 0, -1, 2}, {9, 0, 1, 4}, {5, -1, 0, 9}}
	for _, c := range cases {
		if got := gridSelection(c.from, c.dx, c.dy, 10); got != c.want {
			t.Fatalf("from %d direction %d,%d got %d want %d", c.from, c.dx, c.dy, got, c.want)
		}
	}
	if gridSelection(0, -1, 1, 1) != 0 {
		t.Fatal("single-character roster must remain usable")
	}
}
