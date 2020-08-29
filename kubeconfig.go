package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

type ContextsConfig struct {
	Contexts []ContextsDetails
}

type ContextsDetails struct {
	Name string `yaml:"name"`
}

var Contexts ContextsConfig

func getKubeConfig() (string, error) {
	var kubeConfig string

	if kubeConfig = os.Getenv("KUBECONFIG"); kubeConfig != "" {
		return kubeConfig, nil
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("cannot get user home directory, err: %v", err)
	}

	kubeConfig = filepath.Join(home, ".kube/config")

	return kubeConfig, nil
}

func getKubeConfigContexts(env string) error {
	var TempContext ContextsConfig
	for _, file := range getKubeConfigFiles(env) {
		cfg, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("cannot read file %s, err: %v", file, err)
		}
		if err := yaml.Unmarshal(cfg, &TempContext); err != nil {
			return fmt.Errorf("cannot unmarshal kubeConfig from file %s, err: %v", file, err)
		}
		Contexts.Contexts = append(Contexts.Contexts, TempContext.Contexts...)

	}
	return nil
}

func createLocalKubeConfig(subDir string) (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("cannot get user home directory, err: %v", err)
	}
	appDir := filepath.Join(home, subDir)
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		if err = os.Mkdir(appDir, 0755); err != nil {
			return "", fmt.Errorf("cannot create app dir %s, err: %v", appDir, err)
		}
	}
	return appDir, nil
}

func getFilteredContexts() ([]ContextsDetails, error) {
	var filteredContexts []ContextsDetails
	if len(contextsFlag) == 0 {
		return Contexts.Contexts, nil
	}
	for _, item := range contextsFlag {
		regexpCompiled, err := regexp.Compile(item)
		if err != nil {
			return nil, fmt.Errorf("cannot compile regexp %s, err: %v", item, err)
		}
		for _, ctx := range Contexts.Contexts {
			if regexpCompiled.MatchString(ctx.Name) {
				filteredContexts = append(filteredContexts, ctx)
			}
		}
	}
	return uniqContexts(filteredContexts), nil
}

func uniqContexts(ctx []ContextsDetails) []ContextsDetails {
	var uniqContexts []ContextsDetails
	encountered := map[ContextsDetails]bool{}
	for i := range ctx {
		if encountered[ctx[i]] == true {
		} else {
			encountered[ctx[i]] = true
			uniqContexts = append(uniqContexts, ctx[i])
		}
	}
	return uniqContexts
}

func getKubeConfigFiles(kubeConfig string) []string {
	return strings.Split(kubeConfig, ":")
}
