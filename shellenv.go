/*

Note that if your bash dotfiles are "noisy" (e.g., they print out messages or
nonprintable characters), that might cause problems with using `bash -ic 'type
-at <CMD>'` below. For example, I found that my `tabs -4` line in .bashrc was
printing the following:

𝄢 tabs -4 | hexdump -C
00000000  0d 1b 5b 33 67 1b 48 20  20 20 20 1b 48 20 20 20  |..[3g.H    .H   |
00000010  20 1b 48 20 20 20 20 1b  48 20 20 20 20 1b 48 20  | .H    .H    .H |
00000020  20 20 20 1b 48 20 20 20  20 1b 48 20 20 20 20 1b  |   .H    .H    .|
00000030  48 20 20 20 20 1b 48 20  20 20 20 1b 48 20 20 20  |H    .H    .H   |
00000040  20 1b 48 20 20 20 20 1b  48 20 20 20 20 1b 48 20  | .H    .H    .H |

*/

package sol

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"mvdan.cc/sh/v3/syntax"
)

//go:embed std-cmds/cmds.txt
var stdCmds []byte
var Cmds map[string]string
var nonstdCmds []string
var nonstdCmdDefs map[string]string

type shellEnv struct {
	Aliases   []string
	AliasDefs map[string]string
	Keywords  []string
	Funcs     []string
	FuncDefs  map[string]string
	Builtins  []string
	Vars      []string
	VarDefs   map[string]string
	Paths     []string
}

type cmdType struct {
	Type string
	Def  string
}

func getCmdTypes(cmd string) ([]*cmdType, error) {
	/*
		-t	output a single word which is one of `alias', `keyword',
		`function', `builtin', `file' or `', if NAME is an alias,
		shell reserved word, shell function, shell builtin, disk file,
		or not found, respectively
	*/

	// Remove leading path from command.
	cmd = filepath.Base(cmd)

	cmdTypes := []*cmdType{}

	if _, ok := env.AliasDefs[cmd]; ok {
		cmdTypes = append(cmdTypes, &cmdType{Type: "alias", Def: env.AliasDefs[cmd]})
		log.Debugln(cmd, "is alias")
	}

	for _, k := range env.Keywords {
		if k == cmd {
			cmdTypes = append(cmdTypes, &cmdType{Type: "keyword", Def: ""})
			log.Debugln(cmd, "is keyword")
		}

	}

	if _, ok := env.FuncDefs[cmd]; ok {
		cmdTypes = append(cmdTypes, &cmdType{Type: "function", Def: env.FuncDefs[cmd]})
		log.Debugln(cmd, "is function")
	}

	for _, b := range env.Builtins {
		if b == cmd {
			cmdTypes = append(cmdTypes, &cmdType{Type: "builtin", Def: ""})
			log.Debugln(cmd, "is builtin")
		}
	}

	for _, path := range env.Paths {
		fullPath := filepath.Join(path, cmd)
		_, err := os.Stat(fullPath)
		if err == nil {
			cmdTypes = append(cmdTypes, &cmdType{Type: "file", Def: fullPath})
			log.Debugln(cmd, "is file")
		}
	}

	if len(cmdTypes) == 0 {
		return nil, fmt.Errorf("command %s not found", cmd)
	}

	stdLocs := []string{
		"/bin",
		"/sbin",
		"/usr/bin",
		"/usr/sbin",
	}

	found := false
	for _, ns := range nonstdCmds {
		if ns == cmd {
			found = true
		}
	}
	if !found {
		nonstdCmds = append(nonstdCmds, cmd)
	}
	for _, ct := range cmdTypes[0:1] {

		if ct.Type == "alias" || ct.Type == "function" {
			line, err := ImplodeSh(ct.Def)
			if err != nil {
				log.Warningf("could not implode shell: %v", err)
			}
			nonstdCmdDefs[cmd] = fmt.Sprintf("# %s is %s: %s", cmd, ct.Type, line)
		} else if ct.Type == "file" {
			stdCmd := false
			for _, sc := range strings.Split(string(stdCmds), "\n") {
				if cmd == sc {
					stdCmd = true
				}
			}
			if !stdCmd {
				nonstdCmdDefs[cmd] = fmt.Sprintf("# nonstd cmd: %s", ct.Def)
			} else {
				stdLoc := false
				for _, sl := range stdLocs {
					if strings.HasPrefix(ct.Def, sl) {
						stdLoc = true
					}
				}
				if !stdLoc {
					nonstdCmdDefs[cmd] = fmt.Sprintf("# nonstd cmd dir: %s", ct.Def)
				}
			}
		}
	}

	return cmdTypes, nil

}

func getVarDefs() (map[string]string, error) {

	varDefs, err := exec.Command("bash", "-c", "set -o posix; set").Output()
	if err != nil {
		return nil, fmt.Errorf("could not get var defs: %w", err)
	}
	parser := syntax.NewParser()
	varDefsParsed, err := parser.Parse(strings.NewReader(string(varDefs)), "")

	vds := map[string]string{}

	syntax.Walk(varDefsParsed, func(node syntax.Node) bool {
		if assign, ok := node.(*syntax.Assign); ok && assign.Name != nil && assign.Value != nil {
			var sb strings.Builder
			printer := syntax.NewPrinter()
			printer.Print(&sb, assign.Value)
			vds[assign.Name.Value] = sb.String()
		}
		return true
	})

	for k, v := range vds {
		vFmt, err := fmtProg(v)
		if err != nil {
			log.Warningf("could not format func: %v", err)
		}
		vds[k] = vFmt
	}

	return vds, nil

}

func getFuncDefs() (map[string]string, error) {

	funcDefs, err := exec.Command("bash", "-ic", "declare -f").Output()
	if err != nil {
		return nil, fmt.Errorf("could not get func defs: %w", err)
	}
	parser := syntax.NewParser()
	funcDefsParsed, err := parser.Parse(strings.NewReader(string(funcDefs)), "")
	if err != nil {
		return nil, fmt.Errorf("could not parse func defs: %w", err)
	}

	fds := map[string]string{}

	syntax.Walk(funcDefsParsed, func(node syntax.Node) bool {
		if fd, ok := node.(*syntax.FuncDecl); ok {
			var sb strings.Builder
			printer := syntax.NewPrinter()
			printer.Print(&sb, fd.Body)
			body := sb.String()
			body = strings.TrimPrefix(body, "{")
			body = strings.TrimSuffix(body, "}")
			fds[fd.Name.Value] = body
		}
		return true
	})
	for k, v := range fds {
		vFmt, err := fmtProg(v)
		if err != nil {
			log.Warningf("could not format func: %v", err)
		}
		fds[k] = vFmt
	}

	return fds, nil
}

func getAliasDefs() (map[string]string, error) {

	aliasDefs, err := exec.Command("bash", "-ic", "alias").Output()
	if err != nil {
		return nil, fmt.Errorf("could not get alias defs: %w", err)
	}
	parser := syntax.NewParser()
	aliasDefsParsed, err := parser.Parse(strings.NewReader(string(aliasDefs)), "")
	if err != nil {
		return nil, fmt.Errorf("could not parse alias defs: %w", err)
	}

	ads := map[string]string{}

	syntax.Walk(aliasDefsParsed, func(node syntax.Node) bool {
		if cmd, ok := node.(*syntax.CallExpr); ok {
			if len(cmd.Args) > 0 && cmd.Args[0].Parts != nil {
				if word, ok := cmd.Args[0].Parts[0].(*syntax.Lit); ok && word.Value == "alias" {

					// if len(cmd.Args) < 2 {
					// 	return ""
					// }

					printer := syntax.NewPrinter()

					aliasName := strings.TrimSuffix(cmd.Args[1].Parts[0].(*syntax.Lit).Value, "=")

					finalStr := ""

					for _, p := range cmd.Args[1].Parts[1:] {
						var ts strings.Builder
						printer.Print(&ts, p)
						tsStr := ts.String()
						if tsStr[0] == '\'' && tsStr[len(tsStr)-1] == '\'' {
							tsStr = strings.TrimPrefix(tsStr, "'")
							tsStr = strings.TrimSuffix(tsStr, "'")
						}
						tsStr = strings.ReplaceAll(tsStr, "\\'", "'")
						finalStr = fmt.Sprintf("%s%s", finalStr, tsStr)
					}
					ads[aliasName] = finalStr
				}
			}
		}
		return true
	})

	for k, v := range ads {
		vFmt, err := fmtProg(v)
		if err != nil {
			log.Warningf("could not format alias: %v", err)
		}
		ads[k] = vFmt
	}

	return ads, nil
}

func getShellEnv() (*shellEnv, error) {
	env := &shellEnv{}

	// Aliases
	aliases, err := exec.Command("bash", "-ic", "compgen -a").Output()
	if err != nil {
		return nil, fmt.Errorf("could not get aliases: %w", err)
	}
	env.Aliases = strings.Split(string(aliases), "\n")

	// AliasDefs
	aliasDefs, err := getAliasDefs()
	if err != nil {
		return nil, fmt.Errorf("could not get alias defs: %w", err)
	}
	env.AliasDefs = aliasDefs

	// Builtins
	builtins, err := exec.Command("bash", "-ic", "compgen -b").Output()
	if err != nil {
		return nil, fmt.Errorf("could not get builtins: %w", err)
	}
	env.Builtins = strings.Split(string(builtins), "\n")

	// Funcs
	funcs, err := exec.Command("bash", "-ic", "compgen -A function").Output()
	if err != nil {
		return nil, fmt.Errorf("could not get funcs: %w", err)
	}
	env.Funcs = strings.Split(string(funcs), "\n")

	// FuncDefs
	funcDefs, err := getFuncDefs()
	if err != nil {
		return nil, fmt.Errorf("could not get func defs: %w", err)
	}
	env.FuncDefs = funcDefs

	// Keywords
	keywords, err := exec.Command("bash", "-ic", "compgen -k").Output()
	if err != nil {
		return nil, fmt.Errorf("could not get keywords: %w", err)
	}
	env.Keywords = strings.Split(string(keywords), "\n")

	// Vars
	vars, err := exec.Command("bash", "-ic", "compgen -v").Output()
	if err != nil {
		return nil, fmt.Errorf("could not get vars: %w", err)
	}
	env.Vars = strings.Split(string(vars), "\n")

	// VarDefs
	varDefs, err := getVarDefs()
	if err != nil {
		return nil, fmt.Errorf("could not get var defs: %w", err)
	}
	env.VarDefs = varDefs

	// Extract a few important items.
	if path, ok := env.VarDefs["PATH"]; ok {
		env.Paths = strings.Split(path, ":")
	}

	return env, nil
}
