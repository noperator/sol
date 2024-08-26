package sol

import (
	"fmt"

	"mvdan.cc/sh/v3/syntax"
)

func ImplodeSh(src string) (string, error) {
	pp, err := parseProg(src)
	if err != nil {
		return "", fmt.Errorf("could not parse program: %w", err)
	}

	// We don't have any "insert" changes like we do in the explode case, so
	// we'll move straight to "replace" changes.

	chgsRpl := []change{}
	var walkErr error
	syntax.Walk(pp, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.CallExpr:
			if x.Args != nil {

				if Cfg.Env {
					_, err := getCmdTypes(getCmdVal(*x))
					if err != nil {
						walkErr = fmt.Errorf("could not get command types: %w", err)
						return false
					}
				}

				// For jq/sh command strings, we don't care in this case if the
				// corresponding cfg values are set (like we do in the explode
				// case) since we're indiscriminately imploding the whole
				// program.

				chgsJq, err := fmtJq(x, true, src)
				if err != nil {
					walkErr = fmt.Errorf("could not determine jq changes: %w", err)
					return false
				}
				for _, chg := range chgsJq {
					chgsRpl = append(chgsRpl, chg)
				}

				chgsSh, err := fmtSh(x, true, src)
				if err != nil {
					walkErr = fmt.Errorf("could not determine shell changes: %w", err)
					return false
				}
				for _, chg := range chgsSh {
					chgsRpl = append(chgsRpl, chg)
				}
			}
		}
		return true
	})
	if walkErr != nil {
		return "", fmt.Errorf("could not walk program: %w", walkErr)
	}

	// Apply _replace_ changes.
	srcRpl := modProg(src, chgsRpl)
	pp, err = parseProg(srcRpl)
	if err != nil {
		return "", fmt.Errorf("could not parse program: %w", err)
	}

	return treeToStr(pp, syntax.SingleLine(true)), nil
}
