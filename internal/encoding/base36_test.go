package encoding

import (
	"math/big"
	"testing"
)

func TestEncodeBase36Fixed(t *testing.T) {
	tests := []struct {
		val   int64
		width int
		want  string
	}{
		{0, 3, "000"},
		{35, 3, "00Z"},
		{36, 3, "010"},
		{1295, 2, "ZZ"}, // 1295 = 35 + 35*36 = "ZZ"
	}
	for _, tc := range tests {
		bi := big.NewInt(tc.val)
		got := EncodeBase36Fixed(bi, tc.width)
		if got != tc.want {
			t.Fatalf("EncodeBase36Fixed(%d,%d) = %q; want %q", tc.val, tc.width, got, tc.want)
		}
	}
}
