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

// Static "Class" for handling static website
type Static struct {
	BoltDB   *bolt.DB
	Template *template.Template
	Port     string
}

// Arguments to use with the template
type TemplateArgument struct {
	Members map[string]storage.Member
	Port    string
	Member  storage.Member
}

// Initialise new static handler
func New(db *bolt.DB, tmpl *template.Template, port string) Static {
	static := Static{BoltDB: db, Template: tmpl, Port: port}
	return static
}

// Common function to print an error
func (static Static) PrintError(writer http.ResponseWriter, status int, message string) {
	writer.WriteHeader(status)
	error := api.Error{
		Code:    status,
		Message: message,
	}
	static.Template.ExecuteTemplate(writer, "error", error)
}

// Create new member
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

// List all the members in the database
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
	static.Template.ExecuteTemplate(writer, "index", arguments)
}

// Delete one member from the database
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

// Change member name
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
