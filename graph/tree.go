package graph

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"github.com/awalterschulze/gographviz"
)

// Node a single node that composes the tree
type Node struct {
	Name			string
	Label    		string
	Parent   		*Node
	Children 		[]*Node
	TimeOfParent   	int64
	TimeOfChildren 	int64

	lock           	sync.RWMutex
	level			int
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
	Root 			*Node
	LeafsLevel  	int

	lock 			sync.RWMutex
	nodesByLevel 	map[int][]*Node

}

// NewTree returns a new instance of Tree
func NewTree(root *Node) *Tree {
	t := new(Tree)
	t.Root = root
	t.Root.Name = root.Name     ////// TODO: Check why this
	t.Root.Label = root.Name    /////
	t.Root.Children = []*Node{}
	t.Root.level = 1
	t.LeafsLevel = 1
	t.nodesByLevel = make(map[int][]*Node)
	t.nodesByLevel[1] = append(t.nodesByLevel[1], t.Root)
	return t
}

func (t *Tree) AddNode(n *Node)  {
	l := n.Parent.level + 1
	n.level = l
	t.nodesByLevel[l] = append(t.nodesByLevel[l], n)

	log.Printf(" Adding node to level: %d, current leafsLevel: %d", l, t.LeafsLevel)
	if l > t.LeafsLevel  {
		t.LeafsLevel = l
	} else {
		log.Printf("level comparison failing")
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

	log.Printf("%s%s", s, n.String())
	for _, c := range n.Children {
		t.String(c, s+s)
	}
}

func (t *Tree) ToJSON(n *Node) string {

	if n == nil {
		return ""
	} else if len(n.Children) == 0 {
		return fmt.Sprintf("{\"text\": {\"name\": \"%s\"}},", n.Name)
	} else {
		s := fmt.Sprintf("{\"text\": {\"name\": \"%s\"}, \"children\" : [", n.Name)
		for i, c := range n.Children {
			aux := t.ToJSON(c)

			if i == (len(n.Children) - 1) {
				aux = strings.TrimRight(aux, ",")
			}

			s += aux
		}
		s += "]}"
		return s
	}
}

//func toString(n *Node) string {
//
//	if n == nil {
//		return ""
//	} else if len(n.Children) == 0 {
//		return fmt.Sprintf("{text: {name: \"%s\"}},", n.String())
//	} else {
//		s := fmt.Sprintf("{text: {name: \"%s\"}, children : [", n.String())
//		for _, c := range n.Children {
//			s+= toString(c)
//		}
//		s+= "]}"
//		return s
//	}
//}
func (t *Tree) ToDOT(names, label string) string {

	log.Printf("%s", names)
	name := "T"

	//if n := strings.Split(names, "-")[0]; n != "" {   //TODO: Check why is failing
	//	name = n
	//}

	g := gographviz.NewGraph()
	if err := g.SetName(name); err != nil {
		panic(err)
	}
	labelCurated := fmt.Sprintf("\"%s\"", label)
	g.AddAttr(name, "label", labelCurated)
	g.AddAttr(name, "labelloc", "t")

	if err := g.SetDir(true); err != nil {
		log.Fatal(err)
	}

	attrs := make(map[string]string)
	attrs["label"] = t.Root.Name

	if err := g.AddNode(name, t.Root.Name, attrs); err != nil {
		log.Fatal(err)
	}

	for _, c := range t.Root.Children {
		toDOT(name, g, t.Root, c)
	}

	log.Printf("Potential: %s", g.String())
	return g.String()
}

func toDOT(name string, g *gographviz.Graph, p *Node, n *Node) {

	if n == nil {
		return
	} else {
		attrs := make(map[string]string)

		if n.Label != "" {
			attrs["label"] = n.Label
		}
		if err := g.AddNode(name, n.Name, attrs); err != nil {
			log.Fatal(err)
		}
		if err := g.AddEdge(p.Name, n.Name, true, nil); err != nil {
			log.Fatal(err)
		}

		if len(n.Children) != 0 {
			for _, c := range n.Children {
				toDOT(name, g, n, c)
			}
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