// Package static handles functions related to the static CRUD website.
package static

import (
	"github.com/boltdb/bolt"
	uuidGenerator "github.com/satori/go.uuid"
	"html/template"
	"net/http"
	"rrcrud/api"
	"rrcrud/storage"
	"strings"
)

// IndexTemplate defines the name for the template to be used with the root page.
// It also defines the template filename (with a .tmpl) attached.
const IndexTemplate = "index"

// ErrorTemplate defines the name for the template to be used with the error page.
// It also defines the template filename (with a .tmpl) attached.
const ErrorTemplate = "error"

// Static "Class" for handling static website
type Static struct {
	BoltDB   *bolt.DB
	Template *template.Template
	Port     string
}

// TemplateArgument defines the arguments to use with the templates
type TemplateArgument struct {
	Members map[string]storage.Member
	Port    string
	Member  storage.Member
}

// New initialises a new static handler
func New(db *bolt.DB, tmpl *template.Template, port string) Static {
	static := Static{BoltDB: db, Template: tmpl, Port: port}
	return static
}

// PrintError is a shared function to return an error through the error template (error.tmpl).
func (static Static) PrintError(writer http.ResponseWriter, status int, message string) {
	writer.WriteHeader(status)
	error := api.Error{
		Code:    status,
		Message: message,
	}
	static.Template.ExecuteTemplate(writer, ErrorTemplate, error)
}

// NewMember created a new member with the given name and a new UUID.
func (static Static) NewMember(writer http.ResponseWriter, request *http.Request) {
	name := strings.TrimSpace(request.FormValue("name"))
	if len(name) <= 0 {
		static.PrintError(writer, http.StatusBadRequest, "No name")
		return
	}
	uuid, err := uuidGenerator.NewV4()
	if err != nil {
		static.PrintError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	member := storage.Member{
		ID:   uuid.String(),
		Name: name,
	}
	err = storage.Update(static.BoltDB, member)
	if err != nil {
		static.PrintError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

// ListMembers lists all the members in the database
func (static Static) ListMembers(writer http.ResponseWriter, request *http.Request) {
	arguments := TemplateArgument{
		Port: static.Port,
	}
	members, err := storage.GetMembers(static.BoltDB)
	if err != nil {
		static.PrintError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	arguments.Members = members
	ids, ok := request.URL.Query()["id"]
	if ok && len(ids[0]) > 0 {
		if member, ok := members[ids[0]]; ok {
			arguments.Member = member
		}
	}
	static.Template.ExecuteTemplate(writer, IndexTemplate, arguments)
}

// DeleteMember deletes one member from the database
func (static Static) DeleteMember(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query()["id"][0]
	if len(id) > 0 {
		err := storage.Delete(static.BoltDB, id)
		if err != nil {
			static.PrintError(writer, http.StatusBadRequest, err.Error())
			return
		}
	}
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}

// UpdateMember replaces the member in the database with the same ID
func (static Static) UpdateMember(writer http.ResponseWriter, request *http.Request) {
	id := strings.TrimSpace(request.FormValue("id"))
	if len(id) > 0 {
		member, err := storage.Get(static.BoltDB, id)
		if err != nil {
			static.PrintError(writer, http.StatusInternalServerError, err.Error())
			return
		}
		if len(member.ID) == 0 {
			static.PrintError(writer, http.StatusBadRequest, "not found")
		}
		name := strings.TrimSpace(request.FormValue("name"))
		if len(name) <= 0 {
			static.PrintError(writer, http.StatusBadRequest, "invalid name")
			return
		}
		member.Name = name
		err = storage.Update(static.BoltDB, member)
		if err != nil {
			static.PrintError(writer, http.StatusInternalServerError, err.Error())
			return
		}
	}
	http.Redirect(writer, request, "/", http.StatusSeeOther)
}
