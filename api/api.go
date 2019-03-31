// Package api provides method to process and serve CRUD api requests.
package api

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	uuidGenerator "github.com/satori/go.uuid"
	"log"
	"net/http"
	"rrcrud/storage"
	"strings"
	"time"
)

// API is the struct with the functions.
type API struct {
	BoltDB *bolt.DB
}

// JSONResponse defines the basic structure of the JSON response.
// It contains a list of members, an optional error and a timestamp.
type JSONResponse struct {
	Members   map[string]storage.Member `json:"members,omitempty"`
	Error     *Error                    `json:"error,omitempty"`
	TimeStamp time.Time                 `json:"timestamp"`
}

// Error defines the JSON structure of the error message.
// It contains a Code integer (same as http status codes) and a Message string.
type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// New generates a new API class using the given BoltDB database for operations.
func New(db *bolt.DB) API {
	api := API{db}
	return api
}

// ListMembers returns a JSON response with all the members in the database.
func (api API) ListMembers(writer http.ResponseWriter, request *http.Request) {
	apiError := Error{}
	members, err := storage.GetMembers(api.BoltDB)
	if err != nil {
		apiError.Code = http.StatusInternalServerError
		apiError.Message = err.Error()
	}
	payload, err := getPayload(members, &apiError)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(payload)
	}
}

// GetMember returns a JSON response for only one member from the Mux id variable.
func (api API) GetMember(writer http.ResponseWriter, response *http.Request) {
	status := http.StatusInternalServerError
	selected := make(map[string]storage.Member)
	vars := mux.Vars(response)
	apiError := Error{}
	if id, ok := vars["id"]; ok {
		member, err := storage.Get(api.BoltDB, id)
		if err != nil {
			status = updateError(&apiError, http.StatusInternalServerError, err.Error())
		} else {
			if len(member.ID) > 0 {
				selected[member.ID] = member
				status = http.StatusOK
			} else {
				status = updateError(
					&apiError,
					http.StatusNotFound,
					fmt.Sprintf("no member with ID %s", vars["id"]),
				)
			}
		}
	} else {
		status = updateError(
			&apiError,
			http.StatusBadRequest,
			fmt.Sprintf("nissing id element in payload"),
		)
	}
	payload, err := getPayload(selected, &apiError)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		writer.Write(payload)
	}

}

// DeleteMember deletes from the database the member specified in the Path id variable
func (api API) DeleteMember(writer http.ResponseWriter, reader *http.Request) {
	status := http.StatusInternalServerError
	selected := make(map[string]storage.Member)
	vars := mux.Vars(reader)
	apiError := Error{}
	if id, ok := vars["id"]; ok {
		member, err := storage.Get(api.BoltDB, id)
		if err != nil {
			status = updateError(&apiError, http.StatusInternalServerError, err.Error())
		} else if len(member.ID) == 0 {
			status = updateError(
				&apiError,
				http.StatusNotFound,
				fmt.Sprintf("no member with ID %s", id),
			)
		} else {
			err := storage.Delete(api.BoltDB, id)
			if err != nil {
				status = updateError(&apiError, http.StatusInternalServerError, err.Error())
			}
			status = http.StatusNoContent
			return
		}
	} else {
		status = updateError(
			&apiError,
			http.StatusBadRequest,
			fmt.Sprintf("missing id field in request"),
		)
	}
	payload, err := getPayload(selected, &apiError)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		writer.Write(payload)
	}

}

// UpdateMember replaces a member in the database with one provided by the JSON payload.
func (api API) UpdateMember(writer http.ResponseWriter, reader *http.Request) {
	status := http.StatusInternalServerError
	members := make(map[string]storage.Member)
	vars := mux.Vars(reader)
	apiError := Error{}
	if id, ok := vars["id"]; ok {
		current, err := storage.Get(api.BoltDB, id)
		if err != nil {
			status = updateError(&apiError, http.StatusInternalServerError, err.Error())
		} else {
			if len(current.ID) > 0 {
				member := storage.Member{}
				err = json.NewDecoder(reader.Body).Decode(&member)
				if err == nil {
					if len(strings.TrimSpace(member.Name)) == 0 {
						status = updateError(&apiError, http.StatusBadRequest, fmt.Sprintf("missing or empty name field in request"))
					} else {
						member.ID = id // We ignore and replace whatever id they might have passed with the payload
						err = storage.Update(api.BoltDB, member)
						if err == nil {
							status = http.StatusOK
							members[id] = member
						} else {
							status = updateError(&apiError, http.StatusInternalServerError, err.Error())
						}
					}
				}
			} else {
				status = updateError(&apiError, http.StatusNotFound,fmt.Sprintf("no member with ID %s", id))
			}
		}
	} else {
		status = updateError(&apiError, http.StatusBadRequest, fmt.Sprintf("missing or empty id field in request"))
	}
	payload, err := getPayload(members, &apiError)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		writer.Write(payload)
	}
}

// NewMember adds a new member to the database with the name from the JSON payload and a new UUID
func (api API) NewMember(writer http.ResponseWriter, request *http.Request) {
	status := http.StatusInternalServerError
	var newMembers = make(map[string]storage.Member)
	var payload []byte
	var apiError = Error{}
	uuid, err := uuidGenerator.NewV4()  // Building with go.mod makes this an error, but the code for UUID 1.2 on GitHub is clear
										// https://godoc.org/github.com/satori/go.uuid#NewV4
										// https://github.com/satori/go.uuid/blob/master/generator.go#L68
	if err != nil {
		status = updateError(&apiError, http.StatusInternalServerError, "failed to generate UUID for member")
	} else {
		member := storage.Member{}
		err := json.NewDecoder(request.Body).Decode(&member)
		if err == nil {
			if len(strings.TrimSpace(member.Name)) == 0 {
				status = updateError(&apiError, http.StatusBadRequest, fmt.Sprintf("missing or empty name field in request"))
			} else {
				id := uuid.String()
				member.ID = id // We ignore whatever ID was specified in the POST request
				storage.Update(api.BoltDB, member)
				newMembers[id] = member
				status = http.StatusOK
			}
		} else {
			status = updateError(
				&apiError,
				http.StatusBadRequest,
				fmt.Sprintf("failed to parse JSON body with error %s", err.Error()),
			)
		}
	}
	payload, err = getPayload(newMembers, &apiError)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Fatalf(err.Error())
	} else {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		writer.Write(payload)
	}
}

// getPayload returns a JSON payload with the elements filled according to the arguments.
func getPayload(members map[string]storage.Member, APIError *Error) ([]byte, error) {
	data := JSONResponse{
		TimeStamp: time.Now().UTC(),
	}
	if APIError != nil && (APIError.Code > 0 || len(APIError.Message) > 0) {
		data.Error = APIError
	} else {
		data.Members = members
	}
	payload, error := json.Marshal(data)
	if error != nil {
		log.Printf("failed to marshal payload %v", data)
		log.Print(error.Error())
	}
	return payload, error
}

func updateError(apiError *Error, code int, message string) int {
    apiError.Code = code
	apiError.Message = message
	return code
}
