package midi

import (
	"fmt"
	"strings"
)

type NodeType int

const (
	Root = NodeType(1 + iota)
	Device
	PortGroup
	Input
	Output
)

var nodeTypeNames = []string{
	"(unknown)",
	"root",
	"device",
	"portGroup",
	"input",
	"output",
}

func (nt NodeType) String() string {
	return nodeTypeNames[nt]
}

type Node struct {
	Type         NodeType
	Manufacturer string
	Model        string
	Name         string
	Metadata     map[string]any

	Children []*Node

	Driver Driver
	Entity Entity
}

func (n *Node) String() string {
	return fmt.Sprintf("%s %s %s (%s)", n.Manufacturer, n.Model, n.Name, n.Type)
}

func (n *Node) Inputs() []*Node {
	return n.Collect(func(n *Node) bool {
		return n.Type == Input
	})
}

func (n *Node) Outputs() []*Node {
	return n.Collect(func(n *Node) bool {
		return n.Type == Output
	})
}

func (n *Node) Collect(pred func(*Node) bool) []*Node {
	out := make([]*Node, 0)

	var rec func(n *Node)
	rec = func(n *Node) {
		if pred(n) {
			out = append(out, n)
		}
		for _, c := range n.Children {
			rec(c)
		}
	}

	rec(n)

	return out
}

func (n *Node) Print() {
	n.printInternal(0)
}

func (n *Node) printInternal(indent int) {
	fmt.Printf("%s[%s] manufacturer=\"%s\" model=\"%s\" name=\"%s\"\n", strings.Repeat("  ", indent), n.Type, n.Manufacturer, n.Model, n.Name)
	for _, c := range n.Children {
		c.printInternal(indent + 1)
	}
}
