package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"

	"github.com/caarlos0/env"
	"github.com/gorilla/mux"
)

type Config struct {
	Limit     []string `env:"LIMIT" envSeparator:","`
	Playbooks []string `env:"PLAYBOOKS" envSeparator:","`
	WorkDir   string   `env:"WORK_DIR" envDefault:"/ansible"`
	ApiToken  string   `env:"API_TOKEN"`
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
	if !contains(config.Limit, limit) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	playbook := vars["playbook"]
	if !contains(config.Playbooks, playbook) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	version := r.URL.Query().Get("version")
	if version != "" {
		match, _ := regexp.MatchString("^[a-zA-Z0-9._-]+$", version)
		if !match {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	if config.ApiToken != "" {
		apiTokenHeader := r.Header["X-Api-Token"]
		if len(apiTokenHeader) != 1 || apiTokenHeader[0] != config.ApiToken {
			w.WriteHeader(http.StatusForbidden)
			log.Printf("X-Api-Token %+v header does not match", apiTokenHeader)
			return
		}
	}

	out, err := gitPull()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Git exited with error: %s\n\n%s", err, out)
		return
	}

	out, err = runAnsible(limit, playbook, version)
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

func runAnsible(limit string, playbook string, version string) (string, error) {
	var args []string
	log.Printf("Deploying %s to %s version %s", playbook, limit, version)
	args = append(args, "-e")
	args = append(args, "docker_container_state=deploy")
	if version != "" {
		args = append(args, "-e")
		args = append(args, playbook+"_version="+version)
	}
	args = append(args, "--limit")
	args = append(args, limit)
	args = append(args, "playbooks/"+playbook+".yml")
	return runCommand("ansible-playbook", args...)
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
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
