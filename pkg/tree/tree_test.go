package tree

import "testing"

func TestFindNodeByLevel(t *testing.T) {

	root := Node{Name: "root", Label: "root"}
	nt := NewTree(&root)
	c := &Node{Name: "child1", Label: "child1"}
	c2 := &Node{Name: "child2", Label: "child2"}
	c3 := &Node{Name: "child3", Label: "child3"}
	c4 := &Node{Name: "child4", Label: "child4"}

	root.AddChild(c)
	root.AddChild(c2)

	c3.AddChild(c4)
	root.AddChild(c3)

	nt.AddNode(c)
	nt.AddNode(c2)
	nt.AddNode(c3)
	nt.AddNode(c4)

	strF := "Failed to find node: %s in level:%d"
	if n := nt.FindNodeByLevel("root", 1); n == nil {
		t.Errorf(strF, "root", 1)
	}

	if n := nt.FindNodeByLevel("child1", 2); n == nil {
		t.Errorf(strF, "child1", 2)
	}

	if n := nt.FindNodeByLevel("child2", 2); n == nil {
		t.Errorf(strF, "child2", 2)
	}

	if n := nt.FindNodeByLevel("child3", 2); n == nil {
		t.Errorf(strF, "child3", 2)
	}

	if n := nt.FindNodeByLevel("child4", 3); n == nil {
		t.Errorf(strF, "child4", 3)
	}

}

func TestGetEdges(t *testing.T) {

	root := Node{Name: "root", Label: "root"}
	nt := NewTree(&root)
	c1 := &Node{Name: "child1", Label: "child1"}
	c2 := &Node{Name: "child2", Label: "child2"}
	c3 := &Node{Name: "child3", Label: "child3"}
	c4 := &Node{Name: "child4", Label: "child4"}
	c5 := &Node{Name: "child5", Label: "child5"}

	root.AddChild(c1)
	c1.AddChild(c2)

	c2.AddChild(c3)
	c3.AddChild(c4)
	c3.AddChild(c5)

	nt.AddNode(c1)
	nt.AddNode(c2)
	nt.AddNode(c3)
	nt.AddNode(c4)
	nt.AddNode(c5)

	nt.String(&root, "-")
	r := nt.DFS(&root)
	t.Logf("DFS %v", r)

	for _, v := range r {
		for i := len(v) - 1; i > 0; i-- {
			t.Logf("Edge: %s -- %s", v[i].Label, v[i-1].Label)
		}
	}
}
