package packets

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Handler packets handler
type Handler struct {
	repo Repository
}

// NewHandler creates a new packets Handler
func NewHandler(r Repository) *Handler {
	return &Handler{r}
}

// Save handles POST requests to save new Packet objects
func (h Handler) Save(response http.ResponseWriter, request *http.Request) {
	var packet Packet

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writeErr(response, err)
	}

	err = json.Unmarshal(body, &packet)
	if err != nil {
		writeErr(response, err)
		return
	}

	err = h.repo.Store(packet)
	if err != nil {
		writeErr(response, err)
	}

	err = json.NewEncoder(response).Encode(`{"success":}`)
	if err != nil {
		writeErr(response, err)
	}
}

// GetAll handles GET requests to return a JSON representation of Packets objects
func (h Handler) GetAll(response http.ResponseWriter, request *http.Request) {

	packets, err := h.repo.FindAll()
	if err != nil {
		writeErr(response, err)
	}

	err = json.NewEncoder(response).Encode(packets)
	if err != nil {
		writeErr(response, err)
	}

}

func writeErr(response http.ResponseWriter, err error) {
	log.Printf("error %v", err)
	msg := fmt.Sprintf(`{"message" : "%s"}`, err.Error())
	response.WriteHeader(http.StatusInternalServerError)
	response.Write([]byte(msg))
}
