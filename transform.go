package sol

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/noperator/jqfmt"
	"mvdan.cc/sh/v3/syntax"
)

type change struct {
	Pos int
	End int
	Str string
}

// Returns the modified source after applying specified changes.
func modProg(src string, chgs []change) string {

	// Iterate over positions in reverse order. This is important because going
	// the other direction (forward) would change the length of the source
	// after every change, which would break the Pos/End offsets in each
	// subsequent change we need to make.
	posToIdx := map[int]int{}
	chgPos := []int{}
	for chgIdx, chg := range chgs {
		posToIdx[chg.Pos] = chgIdx
		chgPos = append(chgPos, chg.Pos)
	}
	sort.Ints(chgPos)
	for pos := len(chgPos) - 1; pos >= 0; pos-- {
		idx := posToIdx[chgPos[pos]]
		chg := chgs[idx]
		src = src[:chg.Pos] + chg.Str + src[chg.End:]
	}
	return src
}

// Returns a parsed program (i.e., a syntax tree) that can be walked, etc.
func parseProg(src string) (*syntax.File, error) {
	in := strings.NewReader(src)
	pp, err := syntax.NewParser().Parse(in, "")
	if err != nil {
		return nil, fmt.Errorf("could not parse program: %w", err)
	}
	return pp, nil
}

// TODO: handle multiple options
// opt ... syntax.PrinterOption
// cannot use opt (variable of type []"mvdan.cc/sh/v3/syntax".PrinterOption) as
// "mvdan.cc/sh/v3/syntax".PrinterOption value in argument to syntax.NewPrinter
func treeToStr(pp *syntax.File, opt syntax.PrinterOption) string {

	prtByt := new(bytes.Buffer)
	syntax.NewPrinter(opt).Print(prtByt, pp)
	prtStr := prtByt.String()

	// Remove trailing newline that's printed when a *File is used.
	// - https://pkg.go.dev/mvdan.cc/sh/v3/syntax#Printer.Print
	prtStr = prtStr[:len(prtStr)-1]

	return prtStr
}

// Returns a prettified version of the input program.
func fmtProg(src string) (string, error) {
	pp, err := parseProg(src)
	if err != nil {
		return "", fmt.Errorf("could not parse program: %w", err)
	}
	return treeToStr(pp, syntax.Indent(4)), nil
}

// Make sure that the final line-broken command string has sensible
// indentation. This means that there shouldn't be a "jump" in indentation
// where things are indented at 4 spaces, then 12 spaces (without something
// also indented at 8 spaces).
func normalizeIndents(src string) (string, error) {
	lineSpaces := map[int]int{}

	// Count the number of leading spaces in each line.
	for l, line := range strings.Split(src, "\n") {
		spaceCount := 0
		for _, char := range line {
			if char == ' ' {
				spaceCount++
			} else {
				break
			}
		}
		lineSpaces[l] = spaceCount
	}

	// We'll want to walk back over the command string, starting with the
	// least-indented lines first.
	linesBySpaces := make([]int, 0, len(lineSpaces))
	for l := range lineSpaces {
		linesBySpaces = append(linesBySpaces, l)
	}
	sort.Slice(linesBySpaces, func(i, j int) bool {
		return lineSpaces[linesBySpaces[i]] < lineSpaces[linesBySpaces[j]]
	})

	// We build a map of lines that need to be decremented by 4 spaces...
	last := 0
	dec := map[int]int{}
	for _, l := range linesBySpaces {
		if lineSpaces[l] != last && lineSpaces[l] != last+4 {
			dec[l] = lineSpaces[l] - (last + 4)
		} else {
			last = lineSpaces[l]
		}
	}

	// ...and then apply the decrements.
	srcCln := ""
	for l, ln := range strings.Split(src, "\n") {
		rm := dec[l]
		srcCln += fmt.Sprintf("%s\n", ln[rm:])
	}

	if len(srcCln) <= 1 {
		return srcCln, nil
	} else {
		return srcCln[:len(srcCln)-1], nil
	}
}

func indent(src string, idt int, hang bool) (string, error) {

	// Build indent string.
	idtStr := ""
	for i := 0; i < idt; i++ {
		idtStr += " "
	}

	srcIdt := ""
	first := true
	for _, srcLn := range strings.Split(src, "\n") {

		// Drop blank lines if they made their way in somehow.
		m, err := regexp.MatchString("^ *$", srcLn)
		if err != nil {
			return "", fmt.Errorf("could not match line: %w", err)
		} else if m {
			continue
		}

		// Leave the first line with no added indentation.
		if first && hang {
			srcIdt += fmt.Sprintf("%s\n", srcLn)
			first = false
		} else {

			// Indent all other lines.
			srcIdt += fmt.Sprintf("%s%s\n", idtStr, srcLn)
		}
	}

	if len(srcIdt) <= 1 {
		return srcIdt, nil
	} else {
		return srcIdt[:len(srcIdt)-1], nil
	}
}

func getCmdVal(x syntax.CallExpr) string {
	cmdPart := x.Args[0].Parts[0]
	cmd := ""

	switch c := cmdPart.(type) {

	case *syntax.Lit:
		cmd = c.Value

	case *syntax.SglQuoted:
		cmd = c.Value

	case *syntax.DblQuoted:
		switch c := c.Parts[0].(type) {
		case *syntax.Lit:
			cmd = c.Value
		case *syntax.CmdSubst:
			y := c.Stmts[0].Cmd.(*syntax.CallExpr)
			cmd = getCmdVal(*y)
		}
	}

	return cmd
}

func fmtSh(x *syntax.CallExpr, implode bool, src string) ([]change, error) {

	cmd := getCmdVal(*x)
	atCmdStr := false
	chgs := []change{}
	shArgIdx := 0
	cmdStrArgIdx := 0
	if cmd == "xargs" || cmd == "parallel" {

		// First, try looking for a shell being explicitly invoked.
		// e.g., `xargs bash -c 'echo {}'`
		for a, arg := range x.Args {
			if shArgIdx > 0 {
				break
			}
			part := arg.Parts[0]
			if fmt.Sprintf("%T", part) == "*syntax.Lit" && strings.HasSuffix(part.(*syntax.Lit).Value, "sh") {
				shArgIdx = a
			}
		}

		// Next, try looking for a probable command string.
		// e.g., `xargs 'echo {}'`
		if shArgIdx == 0 {
			for a := len(x.Args) - 1; a >= 0; a-- {
				if cmdStrArgIdx > 0 {
					break
				}
				part := x.Args[a].Parts[0]
				if fmt.Sprintf("%T", part) == "*syntax.SglQuoted" || fmt.Sprintf("%T", part) == "*syntax.DblQuoted" {
					cmdStrArgIdx = a
					atCmdStr = true
				}
			}
		}
	}

	// A non-zero `shArgIdx` indicates that a shell is invoked down the line by
	// another program. For example, in the line below, `shArgIdx == 3` because
	// `bash` is the 3rd argument to `xargs`.
	// ```
	// 0     1   2 3    4  5
	// xargs -0n 2 bash -c 'echo $1'
	// ```

	// TODO: Build a list of common shells instead of matching on suffix. e.g.,
	// bash, csh, dash, ksh, sh, tcsh, zsh

	// TODO: Clean this up.
	if strings.HasSuffix(cmd, "sh") || shArgIdx > 0 || cmdStrArgIdx > 0 {
		var startIdx int
		if shArgIdx > 0 {
			startIdx = shArgIdx + 1
		} else if cmdStrArgIdx > 0 {
			startIdx = cmdStrArgIdx
		}
		for _, arg := range x.Args[startIdx:] {

			// TODO: Do I need to check parts? Does an arg ever have more than
			// one part, or can I just check the first part?
			for _, part := range arg.Parts {

				// If we find `-c`, then we know the next argument is a shell command.
				// TODO: Can we handle something like `-ic` where two args are
				// joined together?
				if fmt.Sprintf("%T", part) == "*syntax.Lit" && part.(*syntax.Lit).Value == "-c" {
					atCmdStr = true
				} else if atCmdStr {
					cmdStr := ""
					lineNumStr := ""
					// TODO: What to do about dollar? Just ignore it?
					syntax.Walk(part, func(node syntax.Node) bool {
						switch x := node.(type) {
						case *syntax.SglQuoted:
							cmdStr = x.Value
							lineNumStr = strings.Split(x.Pos().String(), ":")[0]
						case *syntax.DblQuoted:
							cmdStr = x.Parts[0].(*syntax.Lit).Value
							lineNumStr = strings.Split(x.Parts[0].Pos().String(), ":")[0]
						}
						return true
					})
					lineNum, err := strconv.Atoi(lineNumStr)
					if err != nil {
						err = fmt.Errorf("could not get column value: %w", err)
						return chgs, fmt.Errorf("could not get column value: %w", err)
					}

					line := strings.Split(src, "\n")[lineNum-1]
					spaceCount := 0
					for _, char := range line {
						if char == ' ' {
							spaceCount++
						} else {
							break
						}
					}

					cmdStrMod := ""
					if implode {
						cmdStrMod, err = ImplodeSh(cmdStr)
					} else {
						cmdStrMod, err = ExplodeSh(cmdStr, spaceCount, true)
					}
					if err != nil {
						return chgs, fmt.Errorf("could not format shell: %w", err)
					}
					pos := int(part.Pos().Offset())
					end := int(part.End().Offset())
					chgs = append(chgs, change{pos + 1, end - 1, cmdStrMod})
					break
				}
			}
		}
	}

	return chgs, nil
}

func fmtJq(x *syntax.CallExpr, implode bool, src string) ([]change, error) {

	cmd := getCmdVal(*x)
	chgs := []change{}
	if cmd == "jq" || cmd == "gojq" {
		found := false
		for a := len(x.Args) - 1; a > 0; a-- {
			part := x.Args[a].Parts[0]
			jqStr := ""
			lineNumStr := ""
			switch x := part.(type) {
			case *syntax.SglQuoted:
				found = true
				jqStr = x.Value
				lineNumStr = strings.Split(x.Pos().String(), ":")[0]
			case *syntax.DblQuoted:
				found = true
				jqStr = x.Parts[0].(*syntax.Lit).Value
				lineNumStr = strings.Split(x.Parts[0].Pos().String(), ":")[0]
			}
			if found {

				lineNum, err := strconv.Atoi(lineNumStr)
				if err != nil {
					err = fmt.Errorf("could not get column value: %w", err)
					return chgs, fmt.Errorf("could not get column value: %w", err)
				}

				line := strings.Split(src, "\n")[lineNum-1]
				spaceCount := 0
				for _, char := range line {
					if char == ' ' {
						spaceCount++
					} else {
						break
					}
				}

				jqStrMod, err := jqfmt.DoThing(jqStr, Cfg.JqFmtCfg)
				if err != nil {
					return chgs, fmt.Errorf("could not parse jq: %w", err)
				}

				if !implode {
					jqStrMod, err = indent(jqStrMod, spaceCount+4, true)
					if err != nil {
						return chgs, fmt.Errorf("could not indent query: %w", err)
					}
				}
				pos := int(part.Pos().Offset())
				end := int(part.End().Offset())
				chgs = append(chgs, change{pos + 1, end - 1, jqStrMod})

				// TODO: Parse with gojq and report errors.

				break
			}
		}
	}

	return chgs, nil
}
