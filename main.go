package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
)

const (
	appsubdir = ".mkubectx"
)

type execSummary struct {
	ctx    string
	output []byte
	err    error
}

type Commander interface {
	Command(string, ...string) *exec.Cmd
}

type RealCommander struct{}

func (c RealCommander) Command(command string, args ...string) *exec.Cmd {
	return exec.Command(command, args...)
}

type MockFilesystem interface {
	Stat(string) (os.FileInfo, error)
	Remove(string) error
	WriteFile(string, []byte, os.FileMode) error
}

type RealFilesystem struct{}

func (r RealFilesystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (r RealFilesystem) Remove(name string) error {
	return os.Remove(name)
}

func (r RealFilesystem) WriteFile(name string, data []byte, filemode os.FileMode) error {
	return ioutil.WriteFile(name, data, filemode)
}

func main() {
	cmdArg := getCliFlags()
	kubeConfig, err := getKubeConfig()
	if err != nil {
		log.Fatal(err)
	}
	getKubeConfigContexts(kubeConfig)
	appDir, err := createLocalKubeConfig(appsubdir)
	if err != nil {
		log.Fatal(err)
	}
	execData := make(chan execSummary)
	stopPrinting := make(chan bool)
	defer func() {
		stopPrinting <- true
	}()
	go printCmdOutput(execData, stopPrinting, os.Stdout)
	filteredContexts, err := getFilteredContexts()
	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(len(filteredContexts))
	cmder := RealCommander{}
	fs := RealFilesystem{}
	for _, ctx := range filteredContexts {
		go cmdExec(cmder, fs, ctx.Name, kubeConfig, cmdArg, appDir, execData, &wg)
	}
	wg.Wait()
}
