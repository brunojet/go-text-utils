package encoding

import (
	"math/big"
	"strings"
)

// EncodeBase36Fixed encodes value in base36 with fixed width (padded with 0s).
// Uses big.Int.Text(36) and uppercases/pads to the requested width.
func EncodeBase36Fixed(value *big.Int, width int) string {
	if value == nil {
		return strings.Repeat("0", width)
	}
	s := strings.ToUpper(value.Text(36))
	if len(s) >= width {
		return s
	}
	return strings.Repeat("0", width-len(s)) + s
}
