package ports

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse expands args like "3000", "3000-3010", "8080" into a sorted, deduped
// list of valid TCP ports.
func Parse(args []string) ([]int, error) {
	seen := map[int]struct{}{}
	var out []int
	for _, a := range args {
		if strings.Contains(a, "-") {
			parts := strings.SplitN(a, "-", 2)
			lo, err1 := strconv.Atoi(parts[0])
			hi, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil || lo < 1 || hi > 65535 || lo > hi {
				return nil, fmt.Errorf("invalid port range %q", a)
			}
			for p := lo; p <= hi; p++ {
				if _, ok := seen[p]; ok {
					continue
				}
				seen[p] = struct{}{}
				out = append(out, p)
			}
			continue
		}
		p, err := strconv.Atoi(a)
		if err != nil || p < 1 || p > 65535 {
			return nil, fmt.Errorf("invalid port %q", a)
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out, nil
}
