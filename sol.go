package sol

import (
	"fmt"

	"github.com/noperator/jqfmt"
)

type SolCfg struct {
	Args      bool
	BinCmd    bool
	Clause    bool
	CmdSubst  bool
	Env       bool
	ProcSubst bool
	Redir     bool

	MaxWidth int
	OneLine  bool
	Sh       bool
	Jq       bool
	JqFmtCfg jqfmt.JqFmtCfg
	// JqFuncs []string
}

var Cfg SolCfg

var env *shellEnv

func Format(src string) (string, error) {

	var err error
	Cmds = make(map[string]string)

	if Cfg.Env {

		nonstdCmds = []string{}
		nonstdCmdDefs = map[string]string{}

		env, err = getShellEnv()
		if err != nil {
			return "", fmt.Errorf("could not get shell environment: %v", err)
		}
	}

	srcFmt, err := fmtProg(src)
	if err != nil {
		return "", fmt.Errorf("could not format program: %v", err)
	}

	// First, implode program.
	srcMod, err := ImplodeSh(srcFmt)
	if err != nil {
		return "", fmt.Errorf("could not implode shell: %v", err)
	}

	// If not one-line mode, explode program.
	if !Cfg.OneLine {
		srcMod, err = ExplodeSh(srcMod, 0, true)
		if err != nil {
			return "", fmt.Errorf("could not explode shell: %v", err)
		}
	}

	// Prettify modified program.
	srcModFmt, err := fmtProg(srcMod)
	if err != nil {
		return "", fmt.Errorf("could not format program: %v", err)
	}

	// Normalize indents.
	srcModFmtNml, err := normalizeIndents(srcModFmt)
	if err != nil {
		return "", fmt.Errorf("could not clean up program: %v", err)
	}

	// Prepend non-standard command definitions to formatted program.
	if Cfg.Env {
		for _, cmd := range nonstdCmds {
			if _, ok := nonstdCmdDefs[cmd]; ok {
				srcModFmtNml = nonstdCmdDefs[cmd] + "\n" + srcModFmtNml
			}
		}
	}

	if Cfg.MaxWidth > 0 {
		srcModFmtNmlWid, err := enforceMaxWidth(srcModFmtNml)
		if err != nil {
			return "", fmt.Errorf("could not enforce max width: %v", err)
		}
		return srcModFmtNmlWid, nil
	} else {
		return srcModFmtNml, nil
	}
}
