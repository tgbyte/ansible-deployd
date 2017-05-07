package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/caarlos0/env"
	"github.com/gorilla/mux"
)

type Config struct {
	Limit     []string `env:"LIMIT" envSeparator:","`
	Playbooks []string `env:"PLAYBOOKS" envSeparator:","`
	WorkDir   string   `env:"WORK_DIR" envDefault:"/ansible"`
}

var config Config

func main() {
	os.Exit(run())
}

func run() int {
	err := env.Parse(&config)
	check(err)
	log.Printf("Configuration: %+v", config)

	r := mux.NewRouter()
	r.HandleFunc("/deploy/{limit:[a-zA-Z0-9_-]+}/{playbook:[a-zA-Z0-9_-]+}", DeployHandler).
		Methods(http.MethodPost)

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

	out, err := gitPull()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Git exited with error: %s\n\n%s", err, out)
		return
	}

	out, err = runAnsible(limit, playbook)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Ansible exited with error: %s\n\n%s", err, out)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", out)

}

func gitPull() (string, error) {
	log.Print("Running git pull")
	return runCommand("git", "pull")
}

func runAnsible(limit string, playbook string) (string, error) {
	log.Printf("Deploying %s to %s", playbook, limit)
	return runCommand("ansible-playbook", "-e", "docker_container_state=deploy", "--limit", limit, "playbooks/"+playbook+".yml")
}

func runCommand(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = config.WorkDir
	out, err := cmd.CombinedOutput()

	log.Printf("%s output: %q", name, string(out))
	if err != nil {
		log.Printf("%s exited with error: %s", name, err)
		return string(out), err
	}

	return string(out), nil
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
