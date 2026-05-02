package cli

import (
	"fmt"
	"strconv"
	"strings"
)

// extractPresenceFlags removes bare strict / indefinite / finger-hint tokens from args
// (presence only). Names with '=' or followed by a boolean-looking token
// are rejected.
func extractPresenceFlags(args []string) (rest []string, strict, indefinite, fingerHint bool, err error) {
	strictNames := map[string]struct{}{
		"--strict": {}, "-s": {},
	}
	indefNames := map[string]struct{}{
		"--indefinite": {}, "-i": {},
	}
	fingerNames := map[string]struct{}{
		"--finger-hint": {}, "-f": {},
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		name, _, foundEq := strings.Cut(a, "=")
		if foundEq {
			if _, ok := strictNames[name]; ok {
				return nil, false, false, false, fmt.Errorf("invalid %s: use --strict or -s for strict mode, or omit for non-strict", a)
			}
			if _, ok := indefNames[name]; ok {
				return nil, false, false, false, fmt.Errorf("invalid %s: use --indefinite or -i for indefinite mode, or omit", a)
			}
			if _, ok := fingerNames[name]; ok {
				return nil, false, false, false, fmt.Errorf("invalid %s: use --finger-hint or -f, or omit", a)
			}
			rest = append(rest, a)
			continue
		}
		if _, ok := strictNames[a]; ok {
			if err := rejectBoolLiteralAfterPresenceFlag(args, i, a); err != nil {
				return nil, false, false, false, err
			}
			strict = true
			continue
		}
		if _, ok := indefNames[a]; ok {
			if err := rejectBoolLiteralAfterPresenceFlag(args, i, a); err != nil {
				return nil, false, false, false, err
			}
			indefinite = true
			continue
		}
		if _, ok := fingerNames[a]; ok {
			if err := rejectBoolLiteralAfterPresenceFlag(args, i, a); err != nil {
				return nil, false, false, false, err
			}
			fingerHint = true
			continue
		}
		rest = append(rest, a)
	}
	return rest, strict, indefinite, fingerHint, nil
}

func rejectBoolLiteralAfterPresenceFlag(args []string, i int, flagToken string) error {
	if i+1 >= len(args) {
		return nil
	}
	next := args[i+1]
	if strings.HasPrefix(next, "-") {
		return nil
	}
	if _, perr := strconv.ParseBool(next); perr == nil {
		return fmt.Errorf("invalid: do not pass %q after %s; use the flag alone, or omit", next, flagToken)
	}
	return nil
}
