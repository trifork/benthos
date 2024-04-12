package api

import (
	"bytes"
	"fmt"
	"log"
	"slices"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	yaml "gopkg.in/yaml.v3"
)

func generateGraph(rootNode *yaml.Node) (string, error) {

	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		return "", err
	}
	defer graph.Close()

	graph.SetRankDir(cgraph.LRRank) // Set orientation from left to right

	counter := 0

	addNodesAndEdges(graph, nil, rootNode, &counter, rootNode.Kind)

	var buf bytes.Buffer
	if err := g.Render(graph, "dot", &buf); err != nil {
		//t.log.Fatal(err)
		return "", err
	}

	return buf.String(), nil
}
func addNodesAndEdges(g *cgraph.Graph, parent *cgraph.Node, node *yaml.Node, counter *int, parentKind yaml.Kind) *cgraph.Node {
	// Increment counter for unique node names
	*counter++
	nodeName := fmt.Sprintf("node%d", *counter)
	var graphNode *cgraph.Node
	var err error

	keyNode := getKeyNode(node)

	if node.Kind == yaml.ScalarNode {
		//return graphNode // we do not care about keyvals
	}
	//skip := false
	// Create a node based on node type and value
	if things := []string{"input", "processors", "output", ""}; !slices.Contains(things, keyNode) && parentKind != yaml.SequenceNode && parent != nil {
		fmt.Printf("skipped: v=%v t=%v k=%v l=%v c=%d\n", keyNode, node.Tag, node.Kind, node.Line, *counter)
		//return graphNode
		//skip = true
	} else {
		fmt.Printf("passed: v=%v t=%v k=%v l=%v c=%d\n", keyNode, node.Tag, node.Kind, node.Line, *counter)
	}

	if parent == nil {
		graphNode, err = g.CreateNode(nodeName)
	} else {
		graphNode, err = g.CreateNode(nodeName)
		g.CreateEdge(fmt.Sprintf("edge%d", *counter), parent, graphNode)
	}

	if err != nil {
		log.Fatalf("Failed to create a node or edge: %v", err)
	}

	label := keyNode
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			keyNode, valueNode := node.Content[i], node.Content[i+1]
			if keyNode.Value == "label" && valueNode.Kind == yaml.ScalarNode {
				label = valueNode.Value // Use Benthos label if available
			}
		}
	}

	graphNode.SetLabel(label)

	// Recurse through children if it's a container node
	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode, yaml.MappingNode:
		for _, child := range node.Content {
			if child.Kind == yaml.SequenceNode {
				//for _, seqChild := range child.Content {
				addNodesAndEdges(g, graphNode, child, counter, node.Kind)
				//}
			} else {
				addNodesAndEdges(g, graphNode, child, counter, node.Kind)
			}
		}
	}

	return graphNode
}

// First value should be the metadata with element name/key
func getKeyNode(node *yaml.Node) string {
	keyNode := node.Value
	if keyNode == "" && node.Kind == yaml.MappingNode {
		if len(node.Content) > 0 { // .Content is pair of keyNode, valueNode
			if node.Content[0].Value != "" {
				keyNode = node.Content[0].Value
			}

		}
	}
	if keyNode == "" {
		switch node.Kind {
		case yaml.AliasNode:
			keyNode = "alias"
		case yaml.MappingNode:
			keyNode = "mapping"
		case yaml.SequenceNode:
			keyNode = "sequence"
		case yaml.ScalarNode:
			keyNode = "scalar"
		case yaml.DocumentNode:
			keyNode = "document"
		default:
			keyNode = "unknown"
		}
	}

	return keyNode
}
