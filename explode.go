package sol

import (
	"fmt"

	"mvdan.cc/sh/v3/syntax"
)

// Apply various transformations to a shell program, and indent the result as
// needed (for use in recursion).
func ExplodeSh(src string, idt int, hang bool) (string, error) {

	// First, we determine "insert" changes (namely, inserting line breaks).
	chgsIns := []change{}
	pp, err := parseProg(src)
	if err != nil {
		return "", fmt.Errorf("could not parse program: %w", err)
	}
	syntax.Walk(pp, func(node syntax.Node) bool {
		switch x := node.(type) {

		case *syntax.BinaryCmd:
			if Cfg.BinCmd {
				pos := int(x.Y.Position.Offset())
				chgsIns = append(chgsIns, change{pos, pos, "\n"})
			}

		case *syntax.ProcSubst:
			if Cfg.ProcSubst {
				pos := int(x.OpPos.Offset())
				chgsIns = append(chgsIns, change{pos, pos, "\\\n"})
				chgsIns = append(chgsIns, change{pos + 2, pos + 2, "\n"})
			}

		case *syntax.CmdSubst:
			if Cfg.CmdSubst {
				pos := int(x.Left.Offset())
				chgsIns = append(chgsIns, change{pos, pos, "\\\n"})
				chgsIns = append(chgsIns, change{pos + 2, pos + 2, "\n"})
			}

		case *syntax.CallExpr:

			if Cfg.Args {
				for _, arg := range x.Args {
					for _, part := range arg.Parts {

						// It's hard to know if argument is simply a flag or an
						// option that takes a value. Easiest thing to do is to
						// break on hyphen. This won't catch all cases, (e.g.,
						// positional arguments), but it'll handle 90% of work
						// of sensibily breaking up arguments.
						if string(src[part.Pos().Offset()]) == "-" {
							pos := int(part.Pos().Offset())
							chgsIns = append(chgsIns, change{pos, pos, "\\\n"})
						}
					}
				}
			}

		case *syntax.ForClause:
			if Cfg.Clause {
				pos := int(x.DoPos.Offset()) + 2
				chgsIns = append(chgsIns, change{pos, pos, "\n"})
			}

		case *syntax.WhileClause:
			if Cfg.Clause {
				pos := int(x.DoPos.Offset()) + 2
				chgsIns = append(chgsIns, change{pos, pos, "\n"})
			}

		case *syntax.IfClause:
			if Cfg.Clause {
				var pos int
				if x.ThenPos.IsValid() {
					pos = int(x.ThenPos.Offset()) + 4
				} else {
					pos = int(x.Position.Offset()) + 4
				}
				chgsIns = append(chgsIns, change{pos, pos, "\n"})
			}

		case *syntax.CaseClause:
			if Cfg.Clause {
				pos := int(x.In.Offset()) + 2
				chgsIns = append(chgsIns, change{pos, pos, "\n"})
			}

		case *syntax.CaseItem:
			if Cfg.Clause {
				pos := int(x.OpPos.Offset())
				chgsIns = append(chgsIns, change{pos, pos, "\n"})
				pos = pos + len(x.Op.String())
				chgsIns = append(chgsIns, change{pos, pos, "\n"})
			}

		case *syntax.Redirect:
			if Cfg.Redir {
				pos := int(x.OpPos.Offset())
				chgsIns = append(chgsIns, change{pos, pos, "\\\n"})
			}

		}
		return true
	})

	// Apply _insert_ changes.
	srcIns, err := fmtProg(modProg(src, chgsIns))
	if err != nil {
		return "", fmt.Errorf("could not format program: %w", err)
	}

	// Next, we determine "replace" changes (e.g., replace a inline command
	// string with a line-broken version).
	chgsRpl := []change{}
	pp, err = parseProg(srcIns)
	if err != nil {
		return "", fmt.Errorf("could not parse program: %w", err)
	}
	var walkErr error
	syntax.Walk(pp, func(node syntax.Node) bool {
		switch x := node.(type) {

		case *syntax.CallExpr:

			if len(x.Args) > 0 {

				if Cfg.Env {
					_, err := getCmdTypes(getCmdVal(*x))
					if err != nil {
						walkErr = fmt.Errorf("could not get command types: %w", err)
						return false
					}
				}

				if Cfg.Jq {
					chgsJq, err := fmtJq(x, false, srcIns)
					if err != nil {
						walkErr = fmt.Errorf("could not determine jq changes: %w", err)
						return false
					}
					for _, chg := range chgsJq {
						chgsRpl = append(chgsRpl, chg)
					}
				}

				if Cfg.Sh {
					chgsSh, err := fmtSh(x, false, srcIns)
					if err != nil {
						walkErr = fmt.Errorf("could not determine shell changes: %w", err)
						return false
					}
					for _, chg := range chgsSh {
						chgsRpl = append(chgsRpl, chg)
					}
				}

			}
		}

		return true
	})
	if walkErr != nil {
		return "", fmt.Errorf("could not walk program: %w", walkErr)
	}

	// Apply _replace_ changes.
	srcRpl, err := fmtProg(modProg(srcIns, chgsRpl))
	if err != nil {
		return "", fmt.Errorf("could not format program: %w", err)
	}

	// Indent.
	srcIdt, err := indent(srcRpl, idt, hang)
	if err != nil {
		return "", fmt.Errorf("could not indent program: %w", err)
	}

	return srcIdt, nil
}
