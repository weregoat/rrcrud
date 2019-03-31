package main

import (
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"path"
	"rrcrud/api"
	"rrcrud/static"
	"rrcrud/storage"
)


// BoltName is the defined filename for the BoltDB file.
const BoltName = "members.db"
// TemplatesGlob is the defined string to use for globbing template files.
const TemplatesGlob = "*.tmpl"

// Main provides the entry point to start the http listener for API and/or static CRUD
func main() {

	var err error
	var port= flag.String("port", "8080", "listening port")
	var boltDir= flag.String("boltdb", "/tmp/", "directory for the members.db bolt database")
	var templatesDir= flag.String("templates", "./templates/", "directory with the .tmpl template files")
	var noAPI= flag.Bool("noapi", false, "do no route the API endpoints")
	var noStatic= flag.Bool("nostatic", false, "do not route the static endpoints")
	flag.Parse()

	if *noAPI && *noStatic {
		log.Fatal("no endpoints to run; quitting")
	}

	db, err := bolt.Open(path.Join(*boltDir, BoltName), 0600, nil)
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
	if !*noStatic {
		templates := template.Must(
			template.ParseFiles(
				path.Join(*templatesDir, fmt.Sprintf("%s.tmpl", static.IndexTemplate)),
				path.Join(*templatesDir, fmt.Sprintf("%s.tmpl", static.ErrorTemplate)),
			),
		)
		staticHandler := static.New(db, templates, *port)
		router.HandleFunc("/", staticHandler.ListMembers).Methods(http.MethodGet)
		router.HandleFunc("/edit", staticHandler.ListMembers).Methods(http.MethodPost)
		router.HandleFunc("/new", staticHandler.NewMember).Methods(http.MethodPost)
		router.HandleFunc("/delete", staticHandler.DeleteMember).Methods(http.MethodPost)
		router.HandleFunc("/update", staticHandler.UpdateMember).Methods(http.MethodPost)
		log.Printf("defined static routes")
	}

	// API
	if !*noAPI {
		apiHandler := api.New(db)
		router.HandleFunc("/api/members/", apiHandler.ListMembers).Methods(http.MethodGet)
		router.HandleFunc("/api/member/{id}", apiHandler.GetMember).Methods(http.MethodGet)
		router.HandleFunc("/api/member/{id}", apiHandler.UpdateMember).Methods(http.MethodPut)
		router.HandleFunc("/api/member/{id}", apiHandler.DeleteMember).Methods(http.MethodDelete)
		router.HandleFunc("/api/member/", apiHandler.NewMember).Methods(http.MethodPost)
		log.Printf("defined API routes")
	}

	log.Printf("listening and serving from port %s", *port)
	http.ListenAndServe(fmt.Sprintf(":%s", *port), router)


}
