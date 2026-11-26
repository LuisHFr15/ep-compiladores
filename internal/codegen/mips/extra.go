package mips

import (
	"fmt"
	"strings"
)

func sscanTmp(label string, n *int) (int, error) {
	if !strings.HasPrefix(label, "TN") {
		return 0, fmt.Errorf("not a tmp")
	}
	return fmt.Sscanf(label, "TN%d", n)
}
