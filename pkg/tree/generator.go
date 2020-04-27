package tree

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/letitbeat/dp-analyzer/pkg/dot"
	"github.com/letitbeat/dp-analyzer/pkg/packets"
	"github.com/letitbeat/dp-analyzer/pkg/smt"
	"github.com/letitbeat/dp-analyzer/pkg/topology"
)

// FlowTree represents a data-path which a particular packet follows.
type FlowTree struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	SrcIP      string   `json:"src_ip"`
	DstIP      string   `json:"dst_ip"`
	SrcPort    string   `json:"src_port"`
	DstPort    string   `json:"dst_port"`
	Nodes      string   `json:"nodes"`
	NodesImg   string   `json:"nodes_img"`
	CapturedAt int64    `json:"captured_at"`
	Level      int      `json:"level"`
	Edges      [][]Edge `json:"-"`
	IsSat      bool     `json:"is_sat"`
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
	topo  topology.Topology
	props []smt.Property
}

// NewGenerator creates a new Generator object
func NewGenerator(topo topology.Topology, props []smt.Property) *Generator {
	return &Generator{topo, props}
}

// Generate iterates over all packets receive to construct a FlowTree or set of them
func (g *Generator) Generate(packets map[string][]packets.Packet) ([]FlowTree, error) {

	var trees []FlowTree
	for k := range packets {

		start := time.Now()

		log.Printf("Flow Tree for packets with the following payload/id: %s", k)
		sort.Slice(packets[k], func(i, j int) bool {
			//return packets[k][i].CapturedAt.Before(*packets[k][j].CapturedAt)
			return packets[k][i].CapturedAtNano < packets[k][j].CapturedAtNano
		})
		log.Printf("After DB timestamp: %d, %d", packets[k][0].CapturedAt.Unix(), packets[k][0].CapturedAt.UnixNano())

		firstMsg := packets[k][0]
		root := NewNode(g.getConnectedNode(firstMsg.Device), g.getConnectedNode(firstMsg.Device))

		t := NewTree(root)

		log.Printf("Type:%s SrcIP:%v DstIP:%v SrcPort:%v DstPort:%v",
			firstMsg.GetType(), firstMsg.SrcIP, firstMsg.DstIP, firstMsg.SrcPort, firstMsg.DstPort)

		ft := NewFlowTree(k, firstMsg)

		prime := 0

		for _, p := range packets[k] {

			*p.CapturedAt = time.Unix(0, p.CapturedAtNano) // TODO: Move to another place!

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
					n.TimeOfIngress.IsZero() &&
					n.Parent.TimeOfEgress.Before(*p.CapturedAt) {

					n.TimeOfIngress = p.CapturedAt
				} else if n.Parent != nil &&
					!n.TimeOfIngress.IsZero() &&
					n.TimeOfIngress.Before(*p.CapturedAt) {

					var name string
					if d := t.FindNode(connected); d == nil {
						name = connected
					} else {
						name = fmt.Sprintf("%s_%d", connected, prime)
						prime++
					}
					log.Printf("Creating a node name: %s, label: %s, to attach to: %s", name, connected, n.Name)
					nn := NewNode(name, connected)
					nn.TimeOfEgress = p.CapturedAt
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
				c.TimeOfIngress = p.CapturedAt
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
		ft.Edges = t.Edges()

		trees = append(trees, *ft)

		log.Printf("time elapsed to generate: %s", time.Now().Sub(start))
	}

	return trees, nil
}

// Merge merges a grouped set/slice of FlowTree and returns
// a new slice of them merged.
func (g *Generator) Merge(grouped map[string][]FlowTree) ([]FlowTree, error) {

	var trees []FlowTree
	for k := range grouped {
		var toMerge []FlowTree
		var filesToMerge []string
		var merged FlowTree
		i := 1
		edges := make(map[int][]Edge)
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
				for _, e := range ft.Edges {
					edges[i] = e
					i++
				}
			} else {
				for _, e := range ft.Edges {
					edges[i] = e
					i++
				}
				ft.IsSat = g.smt(edges)
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

			merged.IsSat = g.smt(edges)
			trees = append(trees, merged)
		}
	}

	sort.Slice(trees[:], func(i, j int) bool {
		return trees[i].CapturedAt > trees[j].CapturedAt
	})

	return trees, nil
}

type inputParams struct {
	EdgesCount int
	Edges      map[int][]Edge
	Hosts      []string
	Switches   []string
	DataPlane  []Edge
}

func (g *Generator) smt(edges map[int][]Edge) bool {

	funcMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc": func(i int) int {
			return i + 1
		},
	}
	tmpl := template.Must(template.New("smt.tmpl").Funcs(funcMap).ParseFiles("/app/templates/smt.tmpl"))

	var hosts []string
	for _, h := range g.topo.Hosts {
		hosts = append(hosts, h)
	}
	var switches []string
	for _, s := range g.topo.Switches {
		switches = append(switches, s)
	}

	dataPlane := EdgesFromString(g.topo.DOT)

	params := inputParams{len(edges), edges, hosts, switches, dataPlane}

	var tplCompiled bytes.Buffer
	err := tmpl.Execute(&tplCompiled, params)
	if err != nil {
		log.Printf("error generating smt formula %s", err.Error())
	}

	fname := fmt.Sprintf("/app/scripts/%d.z3", time.Now().UnixNano())
	f, err := os.Create(fname)
	os.Chmod(fname, 0755)
	if err != nil {
		log.Printf("error creating file %s", err)
	}
	defer f.Close()

	content := fmt.Sprintf("%s%s", tplCompiled.String(), g.props[0].Text) //TODO: needs to be extended to more than 1
	_, err = f.WriteString(content)
	if err != nil {
		log.Printf("error writing string to file %s", err)
	}

	f.Sync()
	log.Printf("%s", content)
	r := solve(fname)
	log.Printf("%s", r)
	if r == "sat" {
		return true
	}

	return false
}

func solve(file string) string {
	args := []string{"/app/scripts/solver.py"}
	args = append(args, file)

	cmd := exec.Command("/usr/bin/python", args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("error executing cmd: %s, %s", err, stderr.String())
		return ""
	}
	return strings.TrimSpace(out.String())
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
