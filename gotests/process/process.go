package process

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/cweill/gotests"
)

const newFilePerm os.FileMode = 0644

type Options struct {
	OnlyFuncs     string
	ExclFuncs     string
	ExportedFuncs bool
	AllFuncs      bool
	PrintInputs   bool
	WriteOutput   bool
}

func Run(out io.Writer, args []string, opts *Options) {
	if opts == nil {
		opts = &Options{}
	}
	opt := parseOptions(out, opts)
	if opt == nil {
		return
	}
	if len(args) == 0 {
		fmt.Fprintln(out, "Please specify a file or directory containing the source")
		return
	}
	for _, path := range args {
		generateTests(out, path, opts.WriteOutput, opt)
	}
}

func parseOptions(out io.Writer, opt *Options) *gotests.Options {
	if opt.OnlyFuncs == "" && opt.ExclFuncs == "" && !opt.ExportedFuncs && !opt.AllFuncs {
		fmt.Fprintln(out, "Please specify either the -only, -excl, -export, or -all flag")
		return nil
	}
	onlyRE, err := parseRegexp(opt.OnlyFuncs)
	if err != nil {
		fmt.Fprintln(out, "Invalid -only regex:", err)
		return nil
	}
	exclRE, err := parseRegexp(opt.ExclFuncs)
	if err != nil {
		fmt.Fprintln(out, "Invalid -excl regex:", err)
		return nil
	}
	return &gotests.Options{
		Only:        onlyRE,
		Exclude:     exclRE,
		Exported:    opt.ExportedFuncs,
		PrintInputs: opt.PrintInputs,
	}
}

func parseRegexp(s string) (*regexp.Regexp, error) {
	if s == "" {
		return nil, nil
	}
	re, err := regexp.Compile(s)
	if err != nil {
		return nil, err
	}
	return re, nil
}

func generateTests(out io.Writer, path string, writeOutput bool, opt *gotests.Options) {
	gts, err := gotests.GenerateTests(path, opt)
	if err != nil {
		fmt.Fprintln(out, err.Error())
		return
	}
	if len(gts) == 0 {
		fmt.Fprintln(out, "No tests generated for", path)
		return
	}
	for _, t := range gts {
		outputTest(out, t, writeOutput)
	}
}

func outputTest(out io.Writer, t *gotests.GeneratedTest, writeOutput bool) {
	if writeOutput {
		if err := ioutil.WriteFile(t.Path, t.Output, newFilePerm); err != nil {
			fmt.Fprintln(out, err)
			return
		}
	}
	for _, t := range t.Functions {
		fmt.Fprintln(out, "Generated", t.TestName())
	}
	if !writeOutput {
		out.Write(t.Output)
	}
}
