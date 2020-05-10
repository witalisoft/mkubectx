package main

import (
	"log"
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
	go printCmdOutput(execData, stopPrinting)
	defer func() {
		stopPrinting <- true
	}()
	filteredContexts, err := getFilteredContexts()
	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(len(filteredContexts))
	for _, ctx := range filteredContexts {
		go cmdExec(ctx.Name, kubeConfig, cmdArg, appDir, execData, &wg)
	}
	wg.Wait()
}
