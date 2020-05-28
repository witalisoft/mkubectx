package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

type contexts []string

var contextsFlag contexts

func (c *contexts) String() string {
	return fmt.Sprint(*c)
}

func (c *contexts) Set(val string) error {
	if len(*c) > 0 {
		return errors.New("cannot set contexts multiple times")
	}
	args := strings.Split(val, ",")
	for _, item := range args {
		*c = append(*c, item)
	}
	return nil
}

func getCliFlags() []string {
	helpusage := `Usage:
  mkubectx [-contexts|-c ctx1,ctx2,...] command [args...]
    contexts arguments can be passed as regular expression`
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, helpusage)
	}
	flag.Var(&contextsFlag, "contexts", "")
	flag.Var(&contextsFlag, "c", "")
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	return flag.Args()
}
