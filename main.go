package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/letitbeat/dp-analyzer/pkg/db/mongo"
	"github.com/letitbeat/dp-analyzer/pkg/packets"
	"github.com/letitbeat/dp-analyzer/pkg/smt"
	"github.com/letitbeat/dp-analyzer/pkg/topology"
	"github.com/letitbeat/dp-analyzer/pkg/tree"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/app/") // optionally look for config in the working directory
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("error reading config file: %v", err)
	}
}

func main() {

	client, err := mongo.Connect(context.Background(), viper.GetString("db_uri"))
	if err != nil {
		log.Fatalf("error connecting to db, %v", err)
	}
	topoRepo := topology.NewRepository(client)
	topoHandler := topology.NewHandler(topoRepo)

	packetsRepo := packets.NewRepository(client)
	packetsHandler := packets.NewHandler(packetsRepo)

	smtRepo := smt.NewRepository(client)
	smtHandler := smt.NewHandler(smtRepo)

	treeHandler := tree.NewHandler(packetsRepo, topoRepo, smtRepo)

	router := mux.NewRouter()

	router.HandleFunc("/topology", topoHandler.Set).Methods(http.MethodPost)
	router.HandleFunc("/topology", topoHandler.Get).Methods(http.MethodGet)
	router.HandleFunc("/smt", smtHandler.Save).Methods(http.MethodPost)
	router.HandleFunc("/smt", smtHandler.Get).Methods(http.MethodGet)
	router.HandleFunc("/save", packetsHandler.Save).Methods(http.MethodPost)
	router.HandleFunc("/", treeHandler.GetAll).Methods(http.MethodGet)

	log.Printf("listening on port %d", 5000)
	log.Fatal(http.ListenAndServe(":5000", handlers.CORS()(router)))
}
