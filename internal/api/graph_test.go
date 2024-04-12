package api

import (
	"fmt"
	"os"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func parseYAML(yamlStr string) (*yaml.Node, error) {
	var rootNode yaml.Node
	err := yaml.Unmarshal([]byte(yamlStr), &rootNode)
	if err != nil {
		return nil, err
	}
	return &rootNode, nil
}

// TestGenerateGraph checks if the graph generation correctly processes the node structure
func TestGenerateGraph(t *testing.T) {
	yamlStr, err := os.ReadFile("/workspaces/benthos/config/docker.yaml")
	if err != nil {
		t.Errorf("ReadFile returned an error: %v", err)
		return
	}
	node, err := parseYAML(string(yamlStr))
	if err != nil {
		t.Errorf("generateGraph returned an error: %v for:\n %v", err, string(yamlStr))
		return
	}

	graphStr, err := generateGraph(node)
	if err != nil {
		t.Errorf("generateGraph returned an error: %v", err)
		return
	}

	//fmt.Print("yaml: " + string(yamlStr))
	fmt.Print(graphStr)
	//t.Log(graphStr)

	// This is a very basic check. In practice, you might want to validate the structure more rigorously.
}
