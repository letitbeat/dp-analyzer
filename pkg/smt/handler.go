package smt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

	props, err := h.repo.FindAll()
	if err != nil {
		writeErr(response, err)
	}

	var p Property
	if len(props) > 0 {
		p = props[0]
	}
	err = json.NewEncoder(response).Encode(p)
	if err != nil {
		writeErr(response, err)
	}
}

// Save HTTP POST handler which stores the data-plane topology
func (h *Handler) Save(response http.ResponseWriter, request *http.Request) {

	var property Property

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writeErr(response, err)
	}

	err = json.Unmarshal(body, &property)
	if err != nil {
		writeErr(response, err)
		return
	}

	count, _ := h.repo.Count()

	if count > 0 {
		props, err := h.repo.FindAll()
		if err != nil {
			writeErr(response, err)
			return
		}
		property.ID = props[0].ID
		h.repo.Update(property)
	} else {
		err := h.repo.Store(property)
		if err != nil {
			writeErr(response, err)
			return
		}
	}

	err = json.NewEncoder(response).Encode("{\"status\":\"ok\"}")
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
