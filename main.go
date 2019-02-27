package main

import (
	"net/http"
	"github.com/letitbeat/dp-analyzer/db"
	"github.com/letitbeat/dp-analyzer/graph"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"encoding/base64"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type FlowTree struct {
	Id         string `json:"id"`
	Type       string `json:"type" bson:"Type"`
	SrcIP      string `json:"src_ip" bson:"SrcIP"`
	DstIP      string `json:"dst_ip" bson:"DstIP"`
	SrcPort    string `json:"src_port" bson:"SrcPort"`
	DstPort    string `json:"dst_port"`
	Nodes      string `json:"nodes"`
	NodesImg   string `json:"nodes_img"`
	CapturedAt int64  `json:"captured_at"`
	Level      int    `json:"level"`
}

// GetHandler handles the index route
func GetHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	packets, err := db.FindAll()

	if err != nil {
		log.Fatal(err)
		http.Error(w, "Error getting packets from DB",
			http.StatusInternalServerError)
	}

	packetsMap := make(map[string][]db.PacketWrapper)
	for _, p := range packets {

		packetsMap[p.Payload] = append(packetsMap[p.Payload], p)
	}

	t := generateFlowTrees(packetsMap)

	grouped := make(map[string][]FlowTree)

	for _, f := range t {
		tt := time.Unix(0, f.CapturedAt)
		key := fmt.Sprintf("%s%s%s", f.Type, f.DstPort, tt.Format("2006-01-02 15:04:05"))
		grouped[key] = append(grouped[key], f)
	}

	var final []FlowTree    // final list with merged nodes

	for k := range grouped {

		var toMerge []FlowTree
		var filesToMerge []string
		var merged FlowTree

		for _, f := range grouped[k] {

			if f.Level > 2 {

				fname := fmt.Sprintf("%s.dot", f.Id[:len(f.Id)-10])
				err = writeStringToFile(fname, f.Nodes)

				if err != nil {
					log.Fatal(err)
				}
				filesToMerge = append(filesToMerge, fname)
				toMerge = append(toMerge, f)
				merged = f
			} else {
				log.Printf("level: %v",  f.Level)

				final = append(final, f)
			}
		}

		if len(filesToMerge) > 0 {

			fname := fmt.Sprintf("%s.dot", merged.Id[:len(merged.Id)-10])

			log.Printf("to merge: %v", filesToMerge)
			b := mergeDots(filesToMerge)
			str := string(b)

			err = writeStringToFile(fname, str)

			if err != nil {
				log.Fatalf("Error writing merged dot file: %v", err)
			}

			log.Printf("Merged DOT: %s", str)
			merged.Nodes = str
			dotGraph := generateGraph(fname)

			graphEncoded := base64.StdEncoding.EncodeToString(dotGraph)
			merged.NodesImg = graphEncoded

			final = append(final, merged)
		}

	}

	sort.Slice(final[:], func(i, j int) bool {
		return final[i].CapturedAt > final[j].CapturedAt
	})

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(final)

	w.Write(b.Bytes())
}

func mergeDots(files []string) []byte{

	args := []string{"-guv"}
	args = append(args, files...)

	cmd := exec.Command("gvpack", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out.Bytes()
}

// SaveHandler converts post request body to string
func SaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		var packetmap map[string]interface{}

		//var p PacketWrapper
		err := json.NewDecoder(r.Body).Decode(&packetmap)
		if err != nil {
			log.Printf("request: %v", &packetmap)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("%v", &packetmap)

		db.Save(packetmap)

		fmt.Fprint(w, "POST done")
		log.Printf("Saving data from: %s", r.RemoteAddr)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

type Topology struct {
	Hosts    []string `json:"hosts"`
	Switches []string `json:"switches"`
	Links    []string `json:"links"`
	DOT      string   `json:"dot"`
	DOTImg   string   `json:"dot_img"`
}

func SaveTopologyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		var topology Topology //map[string]interface{}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}

		err = json.Unmarshal(body, &topology)
		if err != nil {
			log.Printf("request: %v", topology)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		f, err := os.OpenFile("topology.json", os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		if _, err = f.WriteString(string(body)); err != nil {
			panic(err)
		}

		log.Printf("Received %v", &topology)

		fmt.Fprintf(w, "POST topology")

	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

}

func GetTopologyHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	topo := getTopology()

	err := writeStringToFile("topo.dot", topo.DOT)
	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	bytes := generateGraph("topo.dot")

	str := base64.StdEncoding.EncodeToString(bytes)

	topo.DOTImg = str

	data, err := json.Marshal(topo)

	if err != nil {
		log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(data)
}

type Filter struct {
	Expression string `json: "expression"`
}

func GeneratePacketsHandler(w http.ResponseWriter, r *http.Request) {

	enableCors(&w)

	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		var f Filter

		err = json.Unmarshal(body, &f)

		if err != nil {
			http.Error(w, "Error parsing request",
				http.StatusInternalServerError)
		}

		log.Printf("Filter requested: %+v", f)

		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(f)

		t := getTopology()
		port := 8800

		for i, h := range t.Hosts {

			url := fmt.Sprintf("http://127.0.0.1:%d/generate", port + i)

			req, err := http.NewRequest("POST", url,  b )
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			client := &http.Client{}

			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			resp.Body.Close()

			log.Printf("response Status: %s for host:%s", resp.Status, h)
		}

		d, err := json.Marshal(t.Hosts)
		w.Write(d)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func getTopology() Topology {

	var topology Topology

	jsonData, err := ioutil.ReadFile("topology.json")

	if err != nil {
		log.Fatalf("Error reading topology.json file: %v", err)
	}

	err = json.Unmarshal(jsonData, &topology)

	if err != nil {
		log.Fatalf("Error converting topology to struct: %v", err)
	}

	return topology
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func generateFlowTrees(packets map[string][]db.PacketWrapper) []FlowTree {

	var trees []FlowTree
	for k := range packets {
		log.Printf("Flow Tree for packets with the following payload/id: %s", k)

		//TODO: Move this to DB
		sort.Slice(packets[k], func(i, j int) bool {
			return packets[k][i].CapturedAt < packets[k][j].CapturedAt
		})

		firstMsg := packets[k][0]
		root := graph.Node{Name: getConnectedNode(firstMsg.Device), Label: getConnectedNode(firstMsg.Device)}
		root.TimeOfChildren = 0
		t := graph.NewTree(&root)

		log.Printf("Type:%s SrcIP:%v DstIP:%v SrcPort:%v DstPort:%v",
			getType(firstMsg.Type), firstMsg.SrcIP, firstMsg.DstIP, firstMsg.SrcPort, firstMsg.DstPort)

		ft := FlowTree{
			Id:         k,
			Type:       getType(firstMsg.Type),
			SrcIP:      firstMsg.SrcIP,
			DstIP:      firstMsg.DstIP,
			SrcPort:    firstMsg.SrcPort,
			DstPort:    firstMsg.DstPort,
			CapturedAt: firstMsg.CapturedAt,
		}

		level := 1

		for _, p := range packets[k] {
			log.Printf("%v %v", p.Device, p.CapturedAt)

			log.Printf("Leafs level %d", t.LeafsLevel)
			level = t.LeafsLevel

			step4:

			n := t.FindNodeByLevel(strings.Split(p.Device, "-")[0], level)
			if n != nil {
				l := getConnectedNode(p.Device)

				log.Printf("Node %s TOP: %d  TOC: %d", n, n.TimeOfParent, n.TimeOfChildren)

				if n.Parent != nil &&
					n.Parent.Name == l &&
					n.TimeOfParent == 0 &&
					n.Parent.TimeOfChildren < p.CapturedAt {

					n.TimeOfParent = p.CapturedAt
					log.Printf("Parent %v, %s", n.Parent.Name, l)
				} else if n.Parent.Name != l &&
					n.TimeOfParent != 0 &&
					n.Parent.TimeOfChildren < p.CapturedAt {

					var name string
					if d := t.FindNode(l); d == nil { //TODO: Double check should be avoided
						name = l
					} else {
						name = fmt.Sprintf("%s_prime", l)
					}
					log.Printf("Create a new node: %s to attach to: %s, d: %s", l, n.Name, p.Device)

					nn := graph.Node{Name: name, Label: l}
					n.AddChild(&nn)
					n.TimeOfChildren = p.CapturedAt

					t.AddNode(&nn)
				} else {
					log.Printf("Error: No parent for message: %v %s", p, l)
				}

			} else if nC := t.FindNodeByLevel(getConnectedNode(p.Device), level); nC != nil {

				d := strings.Split(p.Device, "-")[0]

				log.Printf("Creating a node: %s, from d: %s to attach to: %v", d, p.Device, nC.Name)

				nn := graph.Node{Name: d, Label: d}
				nn.TimeOfParent = p.CapturedAt
				nC.AddChild(&nn)

				t.AddNode(&nn)
			} else {
				//the root level is not the current level
				if level > 1 {
					level--
					goto step4
				} else {
					log.Printf("Error: trying to decrease root level")
				}
			}
		}
		t.String(t.Root, "-")
		label := fmt.Sprintf("%s %s", ft.Type, ft.DstPort)
		dotStr := string(t.ToDOT(k, label))

		fname := fmt.Sprintf("%s.dot", k[:len(k)-10])

		err := writeStringToFile(fname, dotStr)
		if err != nil {
			log.Fatal(err)
		}
		dotGraph := generateGraph(fname)

		str := base64.StdEncoding.EncodeToString(dotGraph)

		ft.Nodes = dotStr
		ft.NodesImg = str
		ft.Level = t.LeafsLevel

		trees = append(trees, ft)
	}

	return trees
}

func writeStringToFile(fname string, str string) error {

	f, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err = f.WriteString(str); err != nil {
		return err
	}
	f.Close()

	return nil
}

func generateGraph(file string) []byte {

	cmd := exec.Command("dot", "-Tpng", file)
	//cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out.Bytes()
}

func getType(t float64) string {
	switch t {
	case 0:
		return "TCP"
	case 1:
		return "UDP"
	default:
		return "Unrecognized"
	}
}

func getConnectedNode(d string) string {

	topo := getTopology()

	for _, l := range topo.Links {
		h := strings.Split(l, ":")[0]
		iface := strings.Split(l, ":")[1]
		if d == iface {
			return h
		}
	}
	log.Printf("Not found for %s ", d)
	return "N/A"
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", GetHandler)
	mux.HandleFunc("/save", SaveHandler)
	mux.HandleFunc("/topology", SaveTopologyHandler)
	mux.HandleFunc("/topo", GetTopologyHandler)

	mux.HandleFunc("/generate", GeneratePacketsHandler)

	log.Printf("listening on port %d", 5000)
	log.Fatal(http.ListenAndServe(":5000", mux))
}
