package main

import (
	"fmt"
	"rrcrud/static"
	// "encoding/json"
	// "fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"rrcrud/api"

	// "github.com/satori/go.uuid"
	"html/template"
	"log"
	"rrcrud/storage"
)
import "net/http"

var tmpl = template.Must(template.ParseGlob("templates/*.tmpl"))
var db *bolt.DB
var port = "8080"

func main() {
	var err error
	db, err = bolt.Open("/tmp/members.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.CreateBucketIfNotExists(storage.MembersBucket)
	if err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	// Static
	staticHandler := static.New(db, tmpl, port)
	router.HandleFunc("/", staticHandler.ListMembers).Methods(http.MethodGet)
	router.HandleFunc("/edit", staticHandler.ListMembers).Methods(http.MethodPost)
	router.HandleFunc("/new", staticHandler.NewMember).Methods(http.MethodPost)
	router.HandleFunc("/delete", staticHandler.DeleteMember).Methods(http.MethodPost)
	router.HandleFunc("/update", staticHandler.UpdateMember).Methods(http.MethodPost)

	// API
	apiHandler := api.New(db)
	router.HandleFunc("/api/members/", apiHandler.ListMembers).Methods(http.MethodGet)
	router.HandleFunc("/api/member/{id}", apiHandler.GetMember).Methods(http.MethodGet)
	router.HandleFunc("/api/member/{id}", apiHandler.UpdateMember).Methods(http.MethodPut)
	router.HandleFunc("/api/member/{id}", apiHandler.DeleteMember).Methods(http.MethodDelete)
	router.HandleFunc("/api/member/", apiHandler.NewMember).Methods(http.MethodPost)

	http.ListenAndServe(fmt.Sprintf(":%s", port), router)
}
