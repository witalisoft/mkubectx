package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type TestCommander struct{}

func (c TestCommander) Command(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestOutput", "--"}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_TEST_OUTPUT=1"}
	return cmd
}

type TestFilesystem struct{}

func (r TestFilesystem) Stat(name string) (os.FileInfo, error) {
	var empty os.FileInfo
	return empty, nil
}

func (r TestFilesystem) Remove(name string) error {
	return nil
}

func (r TestFilesystem) WriteFile(name string, data []byte, filemode os.FileMode) error {
	return nil
}

func TestOutput(*testing.T) {
	if os.Getenv("GO_WANT_TEST_OUTPUT") != "1" {
		return
	}
	defer os.Exit(0)
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}
	cmd, args := args[0], args[1:]
	switch cmd {
	case "success":
		fmt.Fprintf(os.Stdout, "KUBECTX="+os.Getenv("KUBECTX")+",KUBECONFIG="+os.Getenv("KUBECONFIG"))
	case "error":
		fmt.Fprintf(os.Stderr, "KUBECTX="+os.Getenv("KUBECTX")+",KUBECONFIG="+os.Getenv("KUBECONFIG"))
		os.Exit(2)
	}
}

func TestCmdExec(t *testing.T) {
	fakecmd := TestCommander{}
	testfs := TestFilesystem{}
	execData := make(chan execSummary)
	var wg sync.WaitGroup
	ctxname := "testctx"
	mainkubeconfig := "config"
	tests := map[string]map[string]string{
		"success": {
			"stdout": "KUBECTX=" + ctxname + ",KUBECONFIG=" + ctxname + ".yaml:" + mainkubeconfig,
			"stderr": "",
		},
		"error": {
			"stdout": "KUBECTX=" + ctxname + ",KUBECONFIG=" + ctxname + ".yaml:" + mainkubeconfig,
			"stderr": "error in executing command, err: exit status 2",
		},
	}
	for name, output := range tests {
		t.Run(name, func(t *testing.T) {
			wg.Add(1)
			go cmdExec(fakecmd, testfs, ctxname, mainkubeconfig, []string{"cmd", name}, "", execData, &wg)
			data := <-execData
			if string(data.output) != "" {
				if string(data.output) != output["stdout"] {
					t.Errorf("On stdout want \"%s\", got \"%s\"", output["stdout"], string(data.output))
				}
			}
			if data.err != nil {
				if fmt.Sprint(data.err) != output["stderr"] {
					t.Errorf("On stderr want \"%s\", got \"%v\"", output["stderr"], data.err)
				}
				if string(data.output) != output["stdout"] {
					t.Errorf("On stdout want \"%s\", got \"%s\"", output["stdout"], string(data.output))
				}
			}
			if data.ctx != ctxname {
				t.Errorf("Context name want \"%s\", got \"%s\"", ctxname, data.ctx)
			}
		})
	}
}

func TestPrinter(t *testing.T) {
	tests := []struct {
		name     string
		input    execSummary
		expected string
	}{
		{
			name: "success",
			input: execSummary{
				ctx:    "testctx",
				output: []byte("sample output"),
				err:    nil,
			},
			expected: "\x1b[92;1mtestctx\n\x1b[0m  sample output\n",
		},
		{
			name: "error",
			input: execSummary{
				ctx:    "testctx",
				output: nil,
				err:    fmt.Errorf("sample error msg"),
			},
			expected: "\x1b[92;1mtestctx\n\x1b[0m  \n\x1b[91;1m error:\n\x1b[0m  sample error msg\n\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			execData := make(chan execSummary)
			stopPrinting := make(chan bool)
			var output bytes.Buffer
			go printCmdOutput(execData, stopPrinting, &output)
			execData <- test.input
			stopPrinting <- true
			MustEqual(t, test.expected, output.String())
		})
	}
}

func MustEqual(t *testing.T, want, got interface{}) {
	t.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("diff (-want +got):\n%s", diff)
	}
}
