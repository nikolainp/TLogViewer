package config

import (
	"bytes"
	"flag"
)

type PrintUsage struct {
	error
	Usage string
}
type PrintVersion struct {
	error
}

type Config struct {
	programName string

	DataPath       string
	ShowReportOnly bool
}

func New(args []string) (obj Config, err error) {
	var isPrintVersion bool

	obj.programName = args[0]
	fsOut := &bytes.Buffer{}

	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(fsOut)
	fs.BoolVar(&isPrintVersion, "v", false, "print version")
	fs.BoolVar(&obj.ShowReportOnly, "r", false, "just show report")

	if err = fs.Parse(args[1:]); err != nil {
		err = PrintUsage{Usage: fsOut.String()}
		return
	}

	if isPrintVersion {
		err = PrintVersion{}
		return
	}

	if fs.NArg() < 1 {
		err = PrintUsage{Usage: fsOut.String()}
		return
	}

	obj.DataPath = fs.Arg(0)

	return
}
