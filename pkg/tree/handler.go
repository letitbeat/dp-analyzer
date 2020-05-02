package tree

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/letitbeat/dp-analyzer/pkg/packets"
	"github.com/letitbeat/dp-analyzer/pkg/smt"
	"github.com/letitbeat/dp-analyzer/pkg/topology"
)

// Handler topology struct
type Handler struct {
	packetsRepo packets.Repository
	topoRepo    topology.Repository
	smtRepo     smt.Repository
}

// NewHandler returns a new FlowTree Handler
func NewHandler(repo packets.Repository, topoRepo topology.Repository, smtRepo smt.Repository) *Handler {
	return &Handler{repo, topoRepo, smtRepo}
}

// GetAll handles HTTP GET requests and returns a JSON representation of
// the generated FlowTrees
func (h *Handler) GetAll(response http.ResponseWriter, request *http.Request) {

	pks, err := h.packetsRepo.FindAll()
	if err != nil {
		writeErr(response, err)
	}

	packetsMap := make(map[string][]packets.Packet)
	for _, p := range pks {
		packetsMap[p.Payload] = append(packetsMap[p.Payload], p)
	}

	topology, err := h.topoRepo.FindAll()
	if err != nil {
		writeErr(response, err)
	}

	props, err := h.smtRepo.FindAll()
	if err != nil {
		writeErr(response, err)
	}

	g := NewGenerator(topology[0], props)

	trees, err := g.Generate(packetsMap)
	if err != nil {
		writeErr(response, err)
	}

	grouped := make(map[string][]FlowTree)

	for _, t := range trees {
		capturedAt := time.Unix(0, t.CapturedAt)
		key := fmt.Sprintf("%s%s%s", t.Type, t.DstPort, capturedAt.Format(time.RFC3339))
		grouped[key] = append(grouped[key], t)
	}

	final, err := g.Merge(grouped)
	if err != nil {
		writeErr(response, err)
	}

	err = json.NewEncoder(response).Encode(final)
	if err != nil {
		writeErr(response, err)
	}
}

func writeErr(response http.ResponseWriter, err error) {
	msg := fmt.Sprintf(`{"message" : "%s"}`, err.Error())
	log.Println("error ", msg)
	response.WriteHeader(http.StatusInternalServerError)
	response.Write([]byte(msg))
}
