package main

import (
	"log"
	"strings"
	"time"
)
import "net/http"

type Member struct {
	ID               string    `json:"id,omitempty"`
	Name             string    `json:"name,omitempty"`
	RegistrationTime time.Time `json:"registration,omitempty"`
}

var members []Member

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/members/", http.StatusSeeOther)
	})
	mux.HandleFunc("/members/", getMembers)
	mux.HandleFunc("/member/", processMember)
	http.ListenAndServe(":8080", mux)
}

func getMembers(writer http.ResponseWriter, request *http.Request) {
	status := http.StatusBadRequest
	if request.Method == http.MethodGet {
		status = http.StatusOK
		log.Printf("All members")
	}
	writer.WriteHeader(status)
}

func processMember(w http.ResponseWriter, r *http.Request) {
	status := http.StatusBadRequest
	path := strings.Split(strings.Trim(strings.ToLower(r.URL.Path), "/"), "/")
	log.Printf("%v", path)
	w.WriteHeader(status)

}


