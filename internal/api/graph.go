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

	//	addNodesAndEdgesOld(graph, nil, rootNode, &counter, rootNode.Kind)
	addNodesAndEdges(graph, nil, rootNode, &counter, "root")

	var buf bytes.Buffer
	if err := g.Render(graph, "dot", &buf); err != nil {
		//t.log.Fatal(err)
		return "", err
	}

	return buf.String(), nil
}
func addNodesAndEdges(g *cgraph.Graph, parent *cgraph.Node, node *yaml.Node, counter *int, label string) *cgraph.Node {
	// Increment counter for unique node names
	*counter++
	nodeName := fmt.Sprintf("node%d_%s_%d", *counter, getKind(node), len(node.Content))
	var graphNode *cgraph.Node
	var err error

	if node.Kind == yaml.ScalarNode && len(node.Content) == 0 {
		return graphNode // we do not care about keyvals
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

	graphNode.SetLabel(label)

	// Recurse through children if it's a container node
	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, child := range node.Content {
			label := child.Value
			if newLabel, ok := getLabel(child); ok {
				label = newLabel
			}
			addNodesAndEdges(g, graphNode, child, counter, label)
		}

	case yaml.MappingNode:

		if len(node.Content) == 0 {
			return nil
		}

		for i := 0; i < len(node.Content); i += 2 {
			keyNode, valueNode := node.Content[i], node.Content[i+1]
			label := keyNode.Value
			if newLabel, ok := getLabel(valueNode); ok {
				label = newLabel
			}

			//check if all children are scalar nodes and they have no children
			skip := true
			for _, child := range valueNode.Content {
				if label == "processors" {
					skip = false
					break
				}
				if child.Kind != yaml.ScalarNode && child.Kind != yaml.MappingNode {
					skip = false
					break
				}
				if len(child.Content) > 0 {
					skip = false
					break
				}
			}
			//skip = false

			//if len(all(node.Content, yaml.ScalarNode)) != len(node.Content) {
			if !skip {
				addNodesAndEdges(g, graphNode, valueNode, counter, label)
			}
			//}

		}
	}

	return graphNode
}

// Use Benthos label if available
func getLabel(node *yaml.Node) (string, bool) {
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			keyNode, valueNode := node.Content[i], node.Content[i+1]
			if keyNode.Value == "label" && valueNode.Kind == yaml.ScalarNode {
				return valueNode.Value, true
			} else {
				return keyNode.Value, true
			}
		}
	}
	return "", false
}

func except(content []*yaml.Node, filter yaml.Kind) []*yaml.Node {
	var pure []*yaml.Node
	for _, child := range content {
		if child.Kind != yaml.ScalarNode {
			pure = append(pure, child)
		}
	}
	return pure
}

func addNodesAndEdgesOld(g *cgraph.Graph, parent *cgraph.Node, node *yaml.Node, counter *int, parentKind yaml.Kind) *cgraph.Node {
	// Increment counter for unique node names
	*counter++
	nodeName := fmt.Sprintf("node%d_%s_%d", *counter, getKind(node), len(node.Content))
	var graphNode *cgraph.Node
	var err error

	keyNode := getKeyNode(node)

	if node.Kind == yaml.ScalarNode && len(node.Content) == 0 {
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
		//if len(all(node.Content, yaml.ScalarNode)) != len(node.Content) {

		for _, child := range node.Content {
			addNodesAndEdgesOld(g, graphNode, child, counter, node.Kind)
		}
		//}
	}

	return graphNode
}

func all(content []*yaml.Node, filter yaml.Kind) []*yaml.Node {
	var pure []*yaml.Node
	for _, child := range content {
		if child.Kind == filter {
			pure = append(pure, child)
		}
	}
	return pure
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

		keyNode = getKind(node)
	}

	return keyNode
}

func getKind(node *yaml.Node) string {
	switch node.Kind {
	case yaml.AliasNode:
		return "alias"
	case yaml.MappingNode:
		return "mapping"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.DocumentNode:
		return "document"
	default:
		return "unknown"
	}

}
