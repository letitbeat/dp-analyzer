package topology

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/letitbeat/dp-analyzer/pkg/dot"
)

// Handler implements topology operations
type Handler struct {
	repo Repository
}

// NewHandler returns a new topology Handler
func NewHandler(repo Repository) *Handler {
	return &Handler{repo}
}

// Get HTTP GET handler which returns the data-plane topology
func (h *Handler) Get(response http.ResponseWriter, request *http.Request) {

	enableCors(&response)

	t, err := h.getTopology()
	if err != nil {
		writeErr(response, err)
	}

	err = json.NewEncoder(response).Encode(t)
	if err != nil {
		writeErr(response, err)
	}
}

// Set HTTP POST handler which stores the data-plane topology
func (h *Handler) Set(response http.ResponseWriter, request *http.Request) {

	var topology Topology

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writeErr(response, err)
	}

	err = json.Unmarshal(body, &topology)
	if err != nil {
		writeErr(response, err)
		return
	}

	dotStr, err := dot.Generate("topo.dot", topology.DOT)
	if err != nil {
		writeErr(response, err)
		return
	}

	topology.DOTImg = dotStr

	count, _ := h.repo.Count()

	if count > 0 {
		topo, err := h.repo.FindAll()
		if err != nil {
			writeErr(response, err)
			return
		}
		topology.ID = topo[0].ID
		h.repo.Update(topology)
	} else {
		err := h.repo.Store(topology)
		if err != nil {
			writeErr(response, err)
			return
		}
	}

	err = json.NewEncoder(response).Encode(`{"success"}`)
	if err != nil {
		writeErr(response, err)
	}
}

func (h Handler) getTopology() (*Topology, error) {
	topos, err := h.repo.FindAll()
	if err != nil {
		return nil, err
	}
	return &topos[0], nil
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func writeErr(response http.ResponseWriter, err error) {
	msg := fmt.Sprintf(`{"message" : "%s"}`, err.Error())
	response.WriteHeader(http.StatusInternalServerError)
	response.Write([]byte(msg))
}
