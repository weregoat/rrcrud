package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"log"
	"time"
)
import "net/http"

type JSONResponse struct {
	Members      map[string]Member `json:"results"`
	Error        *APIError         `json:"error,omitempty"`
	TimeStamp    time.Time         `json:"timestamp"`
}

type APIError struct {
	Code int `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Member struct {
	ID               string    `json:"id,omitempty"`
	Name             string    `json:"name,omitempty"`
	RegistrationTime time.Time `json:"registration,omitempty"`
}

var members map[string]Member

func main() {
	members = make(map[string]Member)
	router := mux.NewRouter()
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/members/", http.StatusSeeOther)
	})
	router.HandleFunc("/members/", getMembers).Methods(http.MethodGet)
	router.HandleFunc("/member/{id}", getMember).Methods(http.MethodGet)
	router.HandleFunc("/member/{id}", updateMember).Methods(http.MethodPut)
	router.HandleFunc("/member/{id}", deleteMember).Methods(http.MethodDelete)
	router.HandleFunc("/member/", newMember).Methods(http.MethodPost)
	http.ListenAndServe(":8080", router)
}

func getMembers(writer http.ResponseWriter, request *http.Request) {
	payload, err := getPayload(members, nil)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(payload)
	}
}

func getMember(writer http.ResponseWriter, reader *http.Request) {
	status := http.StatusInternalServerError
	selected := make(map[string]Member)
	vars := mux.Vars(reader)
	apiError := APIError{}
	if member, ok := members[vars["id"]]; ok {
		selected[member.ID] = member
		status = http.StatusOK
	} else {
		status = http.StatusNotFound
		apiError.Code = status
		apiError.Message = fmt.Sprintf("No member with ID %s", vars["id"])
	}
	payload, err := getPayload(selected, &apiError)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Fatalf(err.Error())
	} else {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		writer.Write(payload)
	}

}


func deleteMember(writer http.ResponseWriter, reader *http.Request) {
	status := http.StatusInternalServerError
	vars := mux.Vars(reader)
	apiError := APIError{}
	if _, ok := members[vars["id"]]; ok {
		delete(members, vars["id"])
		writer.WriteHeader(http.StatusNoContent)
		return
	} else {
		status = http.StatusNotFound
		apiError.Code = status
		apiError.Message = fmt.Sprintf("No member with ID %s", vars["id"])
		selected := make(map[string]Member)
		payload, err := getPayload(selected, &apiError)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			log.Fatalf(err.Error())
		} else {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(status)
			writer.Write(payload)
		}
	}
}

func updateMember(writer http.ResponseWriter, reader *http.Request) {
	status := http.StatusInternalServerError
	selected := make(map[string]Member)
	vars := mux.Vars(reader)
	apiError := APIError{}
	if old, ok := members[vars["id"]]; ok {
		new := Member{}
		err := json.NewDecoder(reader.Body).Decode(&new)
		if err == nil {
			new.ID = old.ID                             // We ignore whatever ID was specified in the POST request
			new.RegistrationTime = old.RegistrationTime // Same with registration time
			members[old.ID] = new
			selected[old.ID] = new
			status = http.StatusOK
		} else {
			status = http.StatusBadRequest
			apiError.Code = status
			apiError.Message = fmt.Sprintf("Failed to parse JSON body with error %s", err.Error())
		}
	} else {
		status = http.StatusNotFound
		apiError.Code = status
		apiError.Message = fmt.Sprintf("No member with ID %s", vars["id"])
	}
	payload, err := getPayload(selected, &apiError)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Fatalf(err.Error())
	} else {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		writer.Write(payload)
	}
}


func newMember(writer http.ResponseWriter, request *http.Request) {
	status := http.StatusInternalServerError
	var newMembers = make(map[string]Member)
	var payload []byte
	var apiError = APIError{}
	uuid, err := uuid.NewV4()
	if err != nil {
		status = http.StatusInternalServerError // Can change this to something else if more appropriate
		apiError.Code = status
		apiError.Message = "failed to generate UUID for member"
	} else {
		id := uuid.String()
		member := Member{}
		err := json.NewDecoder(request.Body).Decode(&member)
		if err == nil {
			member.ID = id                             // We ignore whatever ID was specified in the POST request
			member.RegistrationTime = time.Now().UTC() // Same with registration time
			members[id] = member
			newMembers[id] = member
			status = http.StatusOK
		} else {
			status = http.StatusBadRequest
			apiError.Code = status
			apiError.Message = fmt.Sprintf("Failed to parse JSON body with error %s", err.Error())
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

func getPayload(members map[string]Member, apiError *APIError) ([]byte, error) {
	data := JSONResponse{
		TimeStamp: time.Now().UTC(),
	}
	if apiError != nil && (apiError.Code > 0 || len(apiError.Message) > 0) {
		data.Error = apiError
	} else {
		data.Members = members
	}
	payload, error := json.Marshal(data)
	return payload, error
}

