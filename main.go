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
	defer func() {
		stopPrinting <- true
	}()
	go printCmdOutput(execData, stopPrinting)
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
