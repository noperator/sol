# sol

A de-minifier (formatter, exploder, beautifier) for shell one-liners.

<p align="center">
<img alt="ray" src="https://i.imgur.com/HfU0Y5X.png" width="700px">
</p>

**sol** [ sohl ] _noun_

1. a tool to help you inspect chained shell commands before you **s**hare a **o**ne-**l**iner (or after you receive one)<br>
    _Before I ran `sol`, I had no idea what the h**k that one-liner I got from Oscar was supposed to do._
2. (also **soul**) the spiritual part of humans regarded in its moral aspect<br>
    _You probably don't have a soul—or at least, not a good one—if you share a one-liner with me without cleaning it up with `sol` first._
3. (rude slang) in a hopeless position or situation<br>
    _You're SOL if you think I'm going to try to read your one-liner without using `sol`._
4. an old French coin equal to 12 deniers<br>
    _`sol` is an free open-source project, but I take dollars and even sols as tips._

### Features

- Choose which transformations you want (break on pipe, args, redirect, whatever)
- "Peeks" into stringified commands (think `xargs`, `parallel`) and formats those, too
- Shows you non-standard aliases, functions, files, etc. that you might not have in your shell environment
- Breaks up long jq lines with [jqfmt](https://github.com/noperator/jqfmt) because—let's be honest—they're getting out of hand

### Built with

- https://github.com/mvdan/sh
- https://github.com/noperator/jqfmt

## Getting started

### Install

```bash
go install -v github.com/noperator/sol/cmd/sol@latest
```

### Usage

```
𝄢 sol -h
Usage of sol:
  -a	arguments
  -all
    	all
  -b	binary commands: &&, ||, |, |&
  -c	command substitution: $(), ``
  -e	inspect env to resolve command types
  -f string
    	file
  -j	jq
  -jqarr
    	arrays
  -jqobj
    	objects
  -jqop string
    	operators (comma-separated)
  -l	clauses: case, for, if, while
  -o	one line
  -p	process substitution: <(), >()
  -r	redirect: >, >>, <, <>, <&, >&, >|, <<, <<-, <<<, &>, &>>
  -s	shell strings: xargs, parallel
  -v	verbose
```

#### via CLI

Explode a complex one-liner directly on an interactive shell prompt. Great for iteratively editing a complex command.

![cli](https://i.imgur.com/7ewABCl.gif)

In the example above, I'm using bash in vi mode; I've bound `@` to `sol-func` which calls `sol` with a few preset options.

```
sol-func() {
	local current_line="${READLINE_LINE}"
	READLINE_LINE=$(echo "$current_line" | sol -p -c -b -r -a -s -jqobj -jqarr -jqop comma)
	READLINE_POINT=${#READLINE_LINE}
}
bind -m vi-command -x '"@": sol-func'
```

#### via Vim

Invoke it directly within Vim using visual block mode, a custom keybinding, etc.

![vim](https://i.imgur.com/M709ZFY.gif)

#### via stdin

Alternatively, you can simply pipe a one-liner into standard input.

![stdin](https://i.imgur.com/lkHZ64V.gif)

## Back matter

### See also

- https://github.com/mvdan/sh/issues/80#issuecomment-385705244
- https://github.com/mvdan/sh/issues/690#issuecomment-810568493
- https://github.com/mvdan/sh/discussions/992

### To-do

- [ ] parallelize `exec.Command` calls in `shellenv.go`
- [ ] log better
- [ ] explicitly handle other shell environments besides `bash`
- [ ] fail gracefully when command not found
- [x] auto-break on 80-char width, etc.
- [ ] document API usage
- [ ] add a test for shell environment inspection

### License

This project is licensed under the [MIT License](LICENSE.md).
