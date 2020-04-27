package tree

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/awalterschulze/gographviz"
)

// Node a single node that composes the tree
type Node struct {
	Name          string
	Label         string
	Parent        *Node
	Children      []*Node
	TimeOfIngress *time.Time
	TimeOfEgress  *time.Time

	lock  sync.RWMutex
	Level int
}

// NewNode creates a new tree Node
func NewNode(name, label string) *Node {
	return &Node{
		Name:          name,
		Label:         label,
		TimeOfIngress: &time.Time{},
		TimeOfEgress:  &time.Time{},
	}
}

// String returns a string representation of this tree
func (n *Node) String() string {
	return fmt.Sprintf("%v", n.Label)
}

// SetParent sets the parent node
func (n *Node) SetParent(p *Node) {
	n.Parent = p
}

// AddChild add a new child to the node
func (n *Node) AddChild(c *Node) {
	n.lock.Lock()
	c.SetParent(n)
	n.Children = append(n.Children, c)
	n.lock.Unlock()
}

// Tree represents a rooted tree objected
type Tree struct {
	Root       *Node
	LeafsLevel int

	lock         sync.RWMutex
	nodesByLevel map[int][]*Node
	edges        []*gographviz.Edge
	nodes        []*gographviz.Node
}

// NewTree returns a new instance of Tree
func NewTree(root *Node) *Tree {
	t := new(Tree)
	t.Root = root
	t.Root.Level = 1
	t.LeafsLevel = 1
	t.nodesByLevel = make(map[int][]*Node)
	t.nodesByLevel[1] = append(t.nodesByLevel[1], t.Root)
	return t
}

// AddNode adds a new Node to the tree
func (t *Tree) AddNode(n *Node) {
	l := n.Parent.Level + 1
	n.Level = l
	t.nodesByLevel[l] = append(t.nodesByLevel[l], n)

	if l > t.LeafsLevel {
		t.LeafsLevel = l
	}
}

// FindNodeByLevel finds a new at a given level of the tree
func (t *Tree) FindNodeByLevel(label string, level int) *Node {

	for _, n := range t.nodesByLevel[level] {
		if n.Label == label {
			return n
		}
	}
	return nil
}

// String returns a string representation of the tree
func (t *Tree) String(n *Node, s string) {
	for _, c := range n.Children {
		log.Printf("%s %s", s, c.Label)
		t.String(c, s+s)
	}
}

// DFS returns an array of parent and children nodes
// after traversing the tree using depth-first search algorithm
func (t *Tree) DFS(n *Node) [][]*Node {

	var r [][]*Node
	if len(n.Children) > 0 {
		for _, c := range n.Children {
			childrenList := t.DFS(c)
			for _, v := range childrenList {
				v = append(v, n) // append the current node the final list
				r = append(r, v)
			}
		}
	} else {
		r = append(r, []*Node{n})
	}
	return r
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

	t.edges = g.Edges.Sorted()
	t.nodes = g.Nodes.Sorted()
	log.Printf("EDGES: %v", g.Edges.Sorted())
	//	log.Printf("Nodes: %s", g.Nodes.Sorted())
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

// Edge holds source and destiny data of a node
type Edge struct {
	Src string
	Dst string
}

// Edges returns an array of edges for all the nodes in
// tree
func (t *Tree) Edges() [][]Edge {

	var edges [][]Edge

	paths := t.DFS(t.Root)
	for _, v := range paths {
		var path []Edge
		for i := len(v) - 1; i > 0; i-- {
			path = append(path, Edge{v[i].Label, v[i-1].Label})
		}
		edges = append(edges, path)
	}
	return edges
}

// EdgesFromString return an array of Edges taking as input
// an string in gographiz format
func EdgesFromString(s string) []Edge {
	var edges []Edge

	g, _ := gographviz.Read([]byte(s))
	for _, e := range g.Edges.Sorted() {
		edges = append(edges, Edge{e.Src, e.Dst})
	}
	return edges
}

func (t *Tree) nodeLabel(name string) string {
	for _, n := range t.nodes {
		if n.Name == name {
			return n.Attrs["label"]
		}
	}
	return ""
}

// PrintNodesAtLevel prints all nodes by given level in the tree
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

// FindNode returns the first node that matches its label
func (t *Tree) FindNode(label string) *Node {
	return findNode(t.Root, label)
}

func findNode(n *Node, label string) *Node {
	if n == nil {
		return nil
	}

	if n.Label == label {
		return n
	}

	for _, c := range n.Children {
		r := findNode(c, label)
		if r != nil {
			return r
		}
	}

	return nil
}
