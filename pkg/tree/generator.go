package tree

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/letitbeat/dp-analyzer/pkg/dot"
	"github.com/letitbeat/dp-analyzer/pkg/packets"
	"github.com/letitbeat/dp-analyzer/pkg/topology"
)

// FlowTree represents a data-path which a particular packet follows.
type FlowTree struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	SrcIP      string `json:"src_ip"`
	DstIP      string `json:"dst_ip"`
	SrcPort    string `json:"src_port"`
	DstPort    string `json:"dst_port"`
	Nodes      string `json:"nodes"`
	NodesImg   string `json:"nodes_img"`
	CapturedAt int64  `json:"captured_at"`
	Level      int    `json:"level"`
}

// NewFlowTree creates a new FlowTree with packets metadata
func NewFlowTree(ID string, p packets.Packet) *FlowTree {
	return &FlowTree{
		ID:         ID,
		Type:       p.GetType(),
		SrcIP:      p.SrcIP,
		DstIP:      p.DstIP,
		SrcPort:    p.SrcPort,
		DstPort:    p.DstPort,
		CapturedAt: p.CapturedAt.UnixNano(),
	}
}

// Generator holds the topology to be used by the generation process
type Generator struct {
	topo topology.Topology
}

// NewGenerator creates a new Generator object
func NewGenerator(topo topology.Topology) *Generator {
	return &Generator{topo}
}

// Generate iterates over all packets receive to construct a FlowTree or set of them
func (g *Generator) Generate(packets map[string][]packets.Packet) ([]FlowTree, error) {

	var trees []FlowTree
	for k := range packets {

		start := time.Now()

		log.Printf("Flow Tree for packets with the following payload/id: %s", k)
		sort.Slice(packets[k], func(i, j int) bool {
			return packets[k][i].CapturedAt.Before(*packets[k][j].CapturedAt)
		})

		firstMsg := packets[k][0]
		root := NewNode(g.getConnectedNode(firstMsg.Device), g.getConnectedNode(firstMsg.Device))

		t := NewTree(root)

		log.Printf("Type:%s SrcIP:%v DstIP:%v SrcPort:%v DstPort:%v",
			firstMsg.GetType(), firstMsg.SrcIP, firstMsg.DstIP, firstMsg.SrcPort, firstMsg.DstPort)

		ft := NewFlowTree(k, firstMsg)

		prime := 0

		for _, p := range packets[k] {
			level := t.LeafsLevel

		step4:

			log.Printf("Device %+v, level: %d", p.Device, level)

			n := t.FindNodeByLevel(strings.Split(p.Device, "-")[0], level)
			log.Printf("FindNodeByLevel node:%s, level: %d, found: %+v", strings.Split(p.Device, "-")[0], level, n)

			if n != nil {
				connected := g.getConnectedNode(p.Device)
				log.Printf("Connected Node: %s", connected)

				if n.Parent != nil &&
					n.Parent.Name == connected &&
					n.TimeOfParent.IsZero() &&
					n.Parent.TimeOfChild.Before(*p.CapturedAt) {

					n.TimeOfParent = p.CapturedAt
				} else if n.Parent != nil &&
					!n.TimeOfParent.IsZero() &&
					n.TimeOfParent.Before(*p.CapturedAt) {

					var name string
					if d := t.FindNode(connected); d == nil {
						name = connected
					} else {
						name = fmt.Sprintf("%s_%d", connected, prime)
						prime++
					}
					log.Printf("Creating a node name: %s, label: %s, to attach to: %s", name, connected, n.Name)
					nn := NewNode(name, connected)
					nn.TimeOfChild = p.CapturedAt
					n.AddChild(nn)
					t.AddNode(nn)
				} else {
					//the root level is not the current level
					if level > 1 {
						level--
						goto step4
					}
					log.Print("error: disconnected data-path")
				}

			} else if n := t.FindNodeByLevel(g.getConnectedNode(p.Device), level); n != nil {

				s := strings.Split(p.Device, "-")[0]

				log.Printf("Creating a node: %s, from device: %s to attach to: %s", s, p.Device, n.Name)

				var name string
				if d := t.FindNode(s); d == nil {
					name = s
				} else {
					name = fmt.Sprintf("%s_%d", s, prime)
					prime++
				}

				c := NewNode(name, s)
				c.TimeOfParent = p.CapturedAt
				n.AddChild(c)
				t.AddNode(c)
			} else {
				//the root level is not the current level
				if level > 1 {
					level--
					goto step4
				}
				log.Print("error: disconnected data-path")
			}
		}

		label := fmt.Sprintf("%s %s", ft.Type, ft.DstPort)
		dotStr := string(t.ToDOT(k, label))

		fname := fmt.Sprintf("%s.dot", k[:len(k)-10])

		dotGraph, err := dot.Generate(fname, dotStr)
		if err != nil {
			log.Printf("error generating flow trees, %s", err.Error())
		}

		ft.Nodes = dotStr
		ft.NodesImg = dotGraph
		ft.Level = t.LeafsLevel

		trees = append(trees, *ft)

		log.Printf("time elapsed to generate: %s", time.Now().Sub(start))
	}

	return trees, nil
}

// Merge merges a grouped set/slice of FlowTree and returns
// a new slice of them merged.
func Merge(grouped map[string][]FlowTree) ([]FlowTree, error) {

	var trees []FlowTree

	for k := range grouped {
		var toMerge []FlowTree
		var filesToMerge []string
		var merged FlowTree

		for _, ft := range grouped[k] {

			if ft.Level > 2 {
				fname := fmt.Sprintf("%s.dot", ft.ID[:len(ft.ID)-10])
				err := dot.Write(fname, ft.Nodes)
				if err != nil {
					log.Printf("error writing file %s, %s", fname, err)
				}
				filesToMerge = append(filesToMerge, fname)
				toMerge = append(toMerge, ft)
				merged = ft
			} else {
				trees = append(trees, ft)
			}
		}

		if len(filesToMerge) > 0 {

			fname := fmt.Sprintf("%s.dot", merged.ID[:len(merged.ID)-10])

			mergedStr, err := dot.Merge(filesToMerge)
			if err != nil {
				log.Printf("error merging files, %s", err.Error())
			}

			dotGraph, err := dot.Generate(fname, mergedStr)
			if err != nil {
				log.Printf("error generating dot from merged files, %s", err.Error())
			}

			merged.Nodes = mergedStr
			merged.NodesImg = dotGraph

			trees = append(trees, merged)
		}
	}

	sort.Slice(trees[:], func(i, j int) bool {
		return trees[i].CapturedAt > trees[j].CapturedAt
	})

	return trees, nil
}

func (g *Generator) getConnectedNode(d string) string {

	for _, l := range g.topo.Links {
		h := strings.Split(l, ":")[0]
		iface := strings.Split(l, ":")[1]
		if d == iface {
			return h
		}
	}
	log.Printf("Not found for %s ", d)
	return "N/A"
}
