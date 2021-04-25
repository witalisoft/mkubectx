package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFilteredContexts(t *testing.T) {
	var savedContexts = Contexts
	var savedContextsFlags = contextsFlag
	Contexts = ContextsConfig{
		Contexts: []ContextsDetails{
			{
				Name: "gke_cluster_ci-0",
			},
			{
				Name: "gke_cluster_pro-1",
			},
			{
				Name: "minikube",
			},
		},
	}
	contextsFlag = []string{
		"ci-\\d+",
		"minikube",
	}
	defer func() {
		Contexts = savedContexts
		contextsFlag = savedContextsFlags
	}()
	var expectedContexts = []ContextsDetails{
		{
			Name: "gke_cluster_ci-0",
		},
		{
			Name: "minikube",
		},
	}
	t.Run("filtering contexts", func(t *testing.T) {
		filteredContexts, err := getFilteredContexts()
		if err != nil {
			t.Fatalf("error while filtering contexts, err: %v", err)
		}
		if diff := cmp.Diff(expectedContexts, filteredContexts); diff != "" {
			t.Errorf("diff (-want +got):\n%s", diff)
		}
	})
}
