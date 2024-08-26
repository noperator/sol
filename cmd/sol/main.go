package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/noperator/jqfmt"
	"github.com/noperator/sol"
	log "github.com/sirupsen/logrus"
)

func main() {

	all := flag.Bool("all", false, "all")
	args := flag.Bool("a", false, "arguments")
	binCmd := flag.Bool("b", false, "binary commands: &&, ||, |, |&")
	clause := flag.Bool("l", false, "clauses: case, for, if, while")
	cmdSubst := flag.Bool("c", false, "command substitution: $(), ````")
	jq := flag.Bool("j", false, "jq")
	procSubst := flag.Bool("p", false, "process substitution: <(), >()")
	redir := flag.Bool("r", false, "redirect: >, >>, <, <>, <&, >&, >|, <<, <<-, <<<, &>, &>>")
	shell := flag.Bool("s", false, "shell strings: xargs, parallel")

	env := flag.Bool("e", false, "inspect env to resolve command types")
	oneLine := flag.Bool("o", false, "one line")
	// jqFuncsStr := flag.String("jf", "group_by,select,sort_by,map", "jq functions")
	file := flag.String("f", "", "file")
	verbose := flag.Bool("v", false, "verbose")

	// jqfmt stuff
	// fn = function, op = operator, ar = array, ob = object
	// funcsStr := flag.String("fn", "group_by,select,sort_by,map", "functions")
	// opsStr := flag.String("op", "pipe", "operators")
	// funcsStr := flag.String("jqfn", "", "functions")
	opsStr := flag.String("jqop", "", "operators (comma-separated)")
	obj := flag.Bool("jqobj", false, "objects")
	arr := flag.Bool("jqarr", false, "arrays")
	// noHang := flag.Bool("nh", false, "no hanging indent")

	flag.Parse()

	if *obj || *arr || *opsStr != "" {
		*jq = true
	}

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	if *file == "" {
		*file = "/dev/stdin"
	}

	// jqfmt stuff
	// var funcs []string
	// if *funcsStr == "" {
	// 	funcs = []string{}
	// } else {
	// 	funcs = strings.Split(*funcsStr, ",")
	// }
	var ops []string
	if *opsStr == "" {
		ops = []string{}
	} else {
		ops = strings.Split(*opsStr, ",")
	}
	jqFmtCfg, err := jqfmt.ValidateConfig(jqfmt.JqFmtCfg{
		Arr: *arr,
		// Funcs: funcs,
		Obj:   *obj,
		OneLn: *oneLine,
		Ops:   ops,
	})
	if err != nil {
		log.Fatalf("invalid jqfmt config: %v", err)
	}

	sol.Cfg.JqFmtCfg = jqFmtCfg

	// TODO: Might need to touch this up.
	if *all {
		*args = true
		*binCmd = true
		*clause = true
		*cmdSubst = true
		*jq = true
		*procSubst = true
		*redir = true
		*shell = true
		*env = true
	} else if *oneLine {
		*args = false
		*binCmd = false
		*clause = false
		*cmdSubst = false
		*jq = false
		*procSubst = false
		*redir = false
		*shell = false
	} else if !(*args || *binCmd || *clause || *cmdSubst || *jq || *procSubst || *redir || *shell) {
		*binCmd = true
	}

	// jqFuncs := strings.Split(*jqFuncsStr, ",")

	sol.Cfg = sol.SolCfg{
		Args:      *args,
		BinCmd:    *binCmd,
		CmdSubst:  *cmdSubst,
		ProcSubst: *procSubst,
		Sh:        *shell,
		Jq:        *jq,
		Clause:    *clause,
		Redir:     *redir,
		JqFmtCfg:  jqFmtCfg,
		OneLine:   *oneLine,
		Env:       *env,
	}

	log.Debugf("cfg: %+v\n", sol.Cfg)

	// Read in program.
	srcBytes, err := os.ReadFile(*file)
	if err != nil {
		log.Fatalf("could not read file: %v", err)
	}
	src := string(srcBytes)

	srcFmt, err := sol.Format(src)
	if err != nil {
		log.Fatalf("could not format program: %v", err)
	}

	fmt.Println(srcFmt)

	os.Exit(0)
}
