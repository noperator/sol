package sol

import (
	// "fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/noperator/jqfmt"
)

func TestExplode(t *testing.T) {

	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/args-in.sh", "testdata/args-out.sh"},
		{"testdata/bincmd-and-in.sh", "testdata/bincmd-and-out.sh"},
		{"testdata/bincmd-or-in.sh", "testdata/bincmd-or-out.sh"},
		{"testdata/bincmd-pipe-in.sh", "testdata/bincmd-pipe-out.sh"},
		{"testdata/bincmd-pipestderr-in.sh", "testdata/bincmd-pipestderr-out.sh"},
		{"testdata/clause-case-in.sh", "testdata/clause-case-out.sh"},
		{"testdata/clause-for-in.sh", "testdata/clause-for-out.sh"},
		{"testdata/clause-if-in.sh", "testdata/clause-if-out.sh"},
		{"testdata/clause-while-in.sh", "testdata/clause-while-out.sh"},
		{"testdata/cmdsubst-backtick-in.sh", "testdata/cmdsubst-backtick-out.sh"}, // `` is deprecated, switches to $()
		{"testdata/cmdsubst-paren-in.sh", "testdata/cmdsubst-paren-out.sh"},
		{"testdata/jq_jqarr-in.sh", "testdata/jq_jqarr-out.sh"},
		{"testdata/jq_jqobj-in.sh", "testdata/jq_jqobj-out.sh"},
		{"testdata/jq_jqop-add-in.sh", "testdata/jq_jqop-add-out.sh"},
		{"testdata/jq_jqop-comma-in.sh", "testdata/jq_jqop-comma-out.sh"},
		{"testdata/jq_jqop-pipe-in.sh", "testdata/jq_jqop-pipe-out.sh"},
		{"testdata/procsubst-input-in.sh", "testdata/procsubst-input-out.sh"},
		{"testdata/procsubst-output-in.sh", "testdata/procsubst-output-out.sh"},
		{"testdata/redir-herestring-in.sh", "testdata/redir-herestring-out.sh"},
		{"testdata/redir-stdall-in.sh", "testdata/redir-stdall-out.sh"},
		{"testdata/redir-stdin-in.sh", "testdata/redir-stdin-out.sh"},
		{"testdata/redir-stdout-in.sh", "testdata/redir-stdout-out.sh"},
		{"testdata/sh_args-parallel-in.sh", "testdata/sh_args-parallel-out.sh"},
		{"testdata/sh_bincmd-xargs-in.sh", "testdata/sh_bincmd-xargs-out.sh"},
	}

	for _, c := range cases {

		cfgTypeStr := strings.Split(strings.TrimPrefix(c.inFile, "testdata/"), "-")[0]
		cfgTypes := []string{}

		jqOps := []string{}
		if strings.HasPrefix(cfgTypeStr, "sh") || strings.HasPrefix(cfgTypeStr, "jq") {
			cfgTypes = strings.Split(cfgTypeStr, "_")
			for _, cfgType := range cfgTypes {
				if cfgType == "jqop" {
					jqOpsStr := strings.Split(strings.TrimPrefix(c.inFile, "testdata/"), "-")[1]
					jqOps = strings.Split(jqOpsStr, "-")
				}
			}

		} else {
			cfgTypes = []string{cfgTypeStr}
		}

		inBytes, err := os.ReadFile(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}
		in := string(inBytes)

		wantBytes, err := os.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}
		want := string(wantBytes)

		Cfg = SolCfg{}

		for _, cfgType := range cfgTypes {
			if cfgType == "clause" {
				Cfg.Clause = true
			}
			if cfgType == "bincmd" {
				Cfg.BinCmd = true
			}
			if cfgType == "redir" {
				Cfg.Redir = true
			}
			if cfgType == "cmdsubst" {
				Cfg.CmdSubst = true
			}
			if cfgType == "procsubst" {
				Cfg.ProcSubst = true
			}
			if cfgType == "args" {
				Cfg.Args = true
			}
			if cfgType == "sh" {
				Cfg.Sh = true
			}
			if cfgType == "jq" {
				Cfg.Jq = true
			}
			if cfgType == "jqobj" {
				Cfg.JqFmtCfg.Obj = true
			}
			if cfgType == "jqarr" {
				Cfg.JqFmtCfg.Arr = true
			}
			if cfgType == "jqop" {
				Cfg.JqFmtCfg.Ops = jqOps
			}
		}

		out, err := Format(in)
		if err != nil {
			t.Fatalf("could not format program: %v", err)
		}

		if !reflect.DeepEqual(want, out) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out)
			t.Errorf("%s does not match %s", c.inFile, c.outFile)
		}
	}
}

func TestComplex(t *testing.T) {

	cases := []struct {
		inFile  string
		cfg     SolCfg
		outFile string
	}{
		{
			"testdata/complex-1-in.sh",
			SolCfg{
				Sh:     true,
				BinCmd: true,
				Args:   true,
				Redir:  true,
				Jq:     true,
				JqFmtCfg: jqfmt.JqFmtCfg{
					Obj: true,
					Arr: true,
					Ops: []string{"pipe", "add"},
				},
			},
			"testdata/complex-1-out.sh",
		},
	}

	for _, c := range cases {

		inBytes, err := os.ReadFile(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}
		in := string(inBytes)

		wantBytes, err := os.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}
		want := string(wantBytes)

		Cfg = c.cfg

		out, err := Format(in)
		if err != nil {
			t.Fatalf("could not format program: %v", err)
		}

		if !reflect.DeepEqual(want, out) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out)
			t.Errorf("%s does not match %s", c.inFile, c.outFile)
		}
	}
}
