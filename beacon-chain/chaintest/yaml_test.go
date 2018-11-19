package main

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/go-yaml/yaml"
	"github.com/prysmaticlabs/prysm/beacon-chain/chaintest/backend"
)

// TestYamls tests constructing test cases from yaml file
func TestYamls(t *testing.T) {
	sb, err := backend.NewSimulatedBackend()
	if err != nil {
		t.Fatalf("Could not create backend: %v", err)
	}

	var chainTests []*backend.ChainTest

	files, err := ioutil.ReadDir("./sampletests")
	if err != nil {
		t.Fatalf("Could not read yaml tests directory: %v", err)
	}

	for _, file := range files {
		data, err := ioutil.ReadFile(path.Join("./sampletests", file.Name()))
		if err != nil {
			t.Fatalf("Could not read yaml file: %v", err)
		}
		decoded := &backend.ChainTest{}
		if err := yaml.Unmarshal(data, decoded); err != nil {
			t.Fatalf("Could not unmarshal YAML file into test struct: %v", err)
		}
		chainTests = append(chainTests, decoded)
	}
}
