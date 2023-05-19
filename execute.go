package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

var colorList = []*color.Color{
	color.New(color.FgHiGreen, color.Bold),
	color.New(color.FgHiYellow, color.Bold),
	color.New(color.FgHiBlue, color.Bold),
	color.New(color.FgHiMagenta, color.Bold),
	color.New(color.FgHiCyan, color.Bold),
	color.New(color.FgHiWhite, color.Bold),
}

type KubeConfigCurrentContext struct {
	Currcontext string `yaml:"current-context"`
}

func cmdExec(cmder Commander, fs MockFilesystem, ctx string, kubeConfig string, cmdArg []string, appDir string, execData chan execSummary, wg *sync.WaitGroup) {
	var combinedOutput []byte
	var combinedErrors []string
	var localKubeConfig string = filepath.Join(appDir, sanitizeKubeconfigFile(ctx)+".yaml")

	defer func() {
		var errorOutput error = nil
		if _, err := fs.Stat(localKubeConfig); err == nil {
			if err := fs.Remove(localKubeConfig); err != nil {
				combinedErrors = append(combinedErrors, fmt.Errorf("cannot remove local kube config file %s, err: %v", localKubeConfig, err).Error())
			}
		}
		if len(combinedErrors) > 0 {
			errorOutput = fmt.Errorf(strings.Join(combinedErrors, "\n"))
		}
		execData <- execSummary{
			ctx:    ctx,
			output: combinedOutput,
			err:    errorOutput,
		}
		wg.Done()
	}()
	fileData, err := yaml.Marshal(&KubeConfigCurrentContext{Currcontext: ctx})
	if err != nil {
		combinedErrors = append(combinedErrors, fmt.Errorf("yaml marshal config problem, err: %v", err).Error())
		return
	}
	err = fs.WriteFile(localKubeConfig, fileData, 0644)
	if err != nil {
		combinedErrors = append(combinedErrors, fmt.Errorf("cannot create new file %s, err: %v", localKubeConfig, err).Error())
		return
	}

	cmd := cmder.Command(cmdArg[0], cmdArg[1:]...)
	cmd.Env = append(cmd.Env, append(os.Environ(), "KUBECONFIG="+localKubeConfig+":"+kubeConfig, "KUBECTX="+ctx)...)
	combinedOutput, err = cmd.CombinedOutput()
	if err != nil {
		combinedErrors = append(combinedErrors, fmt.Errorf("error in executing command, err: %v", err).Error())
		return
	}
	return

}

func printCmdOutput(execData chan execSummary, stopPrinting chan bool, output io.Writer) {
	var coloridx int
L:
	for {
		select {
		case data := <-execData:
			for i, line := range strings.Split(string(data.output), "\n") {
				if i == 0 {
					color := colorList[coloridx%len(colorList)]
					color.Fprintf(output, "%s\n", data.ctx)
					coloridx++
				}
				fmt.Fprintf(output, "  %s\n", line)
			}
			if data.err != nil {
				color := color.New(color.FgHiRed, color.Bold)
				color.Fprintf(output, " error:\n")
				fmt.Fprintf(output, "  %v\n\n", data.err)
			}
		case <-stopPrinting:
			break L
		default:
			continue
		}
	}
}

func sanitizeKubeconfigFile(s string) string {
	re := regexp.MustCompile(`[:/]`)
	return re.ReplaceAllString(s, "_")

}
