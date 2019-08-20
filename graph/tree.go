package graph

import (
	"fmt"
	"log"
	"sync"

	"github.com/awalterschulze/gographviz"
)

// Node a single node that composes the tree
type Node struct {
	Name           string
	Label          string
	Parent         *Node
	Children       []*Node
	TimeOfParent   int64
	TimeOfChildren int64

	lock  sync.RWMutex
	level int
}

func (n *Node) String() string {
	return fmt.Sprintf("%v", n.Label)
}

func (n *Node) SetParent(p *Node) {
	n.Parent = p
}

func (n *Node) AddChild(c *Node) {
	n.lock.Lock()
	c.SetParent(n)
	n.Children = append(n.Children, c)
	n.lock.Unlock()
}

type Tree struct {
	Root       *Node
	LeafsLevel int

	lock         sync.RWMutex
	nodesByLevel map[int][]*Node
}

// NewTree returns a new instance of Tree
func NewTree(root *Node) *Tree {
	t := new(Tree)
	t.Root = root
	t.Root.Name = root.Name  ////// TODO: Check why this
	t.Root.Label = root.Name /////
	t.Root.Children = []*Node{}
	t.Root.level = 1
	t.LeafsLevel = 1
	t.nodesByLevel = make(map[int][]*Node)
	t.nodesByLevel[1] = append(t.nodesByLevel[1], t.Root)
	return t
}

func (t *Tree) AddNode(n *Node) {
	l := n.Parent.level + 1
	n.level = l
	t.nodesByLevel[l] = append(t.nodesByLevel[l], n)

	if l > t.LeafsLevel {
		t.LeafsLevel = l
	}
}

func (t *Tree) FindNodeByLevel(label string, level int) *Node {

	for _, n := range t.nodesByLevel[level] {
		if n.Label == label {
			return n
		}
	}
	return nil
}

func (t *Tree) String(n *Node, s string) {
	for _, c := range n.Children {
		t.String(c, s+s)
	}
}

// ToDOT returns a string of the current tree in
// DOT format.
func (t *Tree) ToDOT(names, label string) string {

	log.Printf("%s", names)
	name := "T"

	//if n := strings.Split(names, "-")[0]; n != "" {   //TODO: Check why is failing
	//	name = n
	//}

	g := gographviz.NewGraph()
	if err := g.SetName(name); err != nil {
		log.Println(err)
	}
	labelCurated := fmt.Sprintf("\"%s\"", label)
	g.AddAttr(name, "label", labelCurated)
	g.AddAttr(name, "labelloc", "t")

	if err := g.SetDir(true); err != nil {
		log.Println(err)
	}

	attrs := make(map[string]string)
	attrs["label"] = t.Root.Name

	if err := g.AddNode(name, t.Root.Name, attrs); err != nil {
		log.Println(err)
	}

	for _, c := range t.Root.Children {
		toDOT(name, g, t.Root, c)
	}

	return g.String()
}

func toDOT(name string, g *gographviz.Graph, p *Node, n *Node) {

	if n == nil {
		return
	}
	attrs := make(map[string]string)

	if n.Label != "" {
		attrs["label"] = n.Label
	}
	if err := g.AddNode(name, n.Name, attrs); err != nil {
		log.Println(err)
	}
	if err := g.AddEdge(p.Name, n.Name, true, nil); err != nil {
		log.Println(err)
	}

	if len(n.Children) != 0 {
		for _, c := range n.Children {
			toDOT(name, g, n, c)
		}
	}
}

func PrintNodesAtLevel(n *Node, currentLevel int, level int) {

	if n == nil {
		return
	}

	if currentLevel == level {
		log.Printf(" %s", n.String())
		return
	}

	for _, c := range n.Children {
		PrintNodesAtLevel(c, currentLevel+1, level)
	}
}

func (t *Tree) FindNode(label string) *Node {
	return findNode(t.Root, label)
}

func findNode(n *Node, label string) *Node {
	if n == nil {
		return nil
	}

	if n.Label == label {
		return n
	} else {
		for _, c := range n.Children {
			r := findNode(c, label)
			if r != nil {
				return r
			}
		}
	}
	return nil
}
