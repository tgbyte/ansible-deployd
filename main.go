package main

import (
	"os"
	"fmt"

	"github.com/caarlos0/env"
	"github.com/gorilla/mux"
	"net/http"
	"log"
)

type Config struct {
	Limit     []string      `env:"LIMIT" envSeparator:","`
	Playbooks []string      `env:"PLAYBOOKS" envSeparator:","`
}

var config Config

func main() {
	os.Exit(run())
}

func run() int {
	err := env.Parse(&config)
	check(err)
	fmt.Printf("%+v\n", config)

	r := mux.NewRouter()
	r.HandleFunc("/deploy/{limit:[a-zA-Z0-9_-]+}/{playbook:[a-zA-Z0-9_-]+}", DeployHandler)

	log.Fatal(http.ListenAndServe(":8000", r))

	return 0
}

func DeployHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	limit := vars["limit"]
	if ! contains(config.Limit, limit) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	playbook := vars["playbook"]
	if ! contains(config.Playbooks, playbook) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Limit: %v, Playbook: %v\n", limit, playbook)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

