module github.com/noperator/sol

go 1.21.5

toolchain go1.22.5

require (
	github.com/noperator/jqfmt v0.0.0-20240815185611-8fc6f864c295
	github.com/sirupsen/logrus v1.9.3
	mvdan.cc/sh/v3 v3.5.1
// noperator.dev/jqfmt v0.0.0-00010101000000-000000000000
)

require (
	github.com/itchyny/gojq v0.12.14 // indirect
	github.com/itchyny/timefmt-go v0.1.5 // indirect
	golang.org/x/sys v0.15.0 // indirect
)

// replace noperator.dev/jqfmt => ../jqfmt
