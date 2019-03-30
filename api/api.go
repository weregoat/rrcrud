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
	"time"
)

type API struct {
	BoltDB *bolt.DB
}

type JSONResponse struct {
	Members   map[string]storage.Member `json:"results"`
	Error     *Error                    `json:"error,omitempty"`
	TimeStamp time.Time                 `json:"timestamp"`
}

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func New(db *bolt.DB) API {
	api := API{db}
	return api
}

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

func (api API) GetMember(writer http.ResponseWriter, response *http.Request) {
	status := http.StatusInternalServerError
	selected := make(map[string]storage.Member)
	vars := mux.Vars(response)
	apiError := Error{}
	if id, ok := vars["id"]; ok {
		member, error := storage.Get(api.BoltDB, id)
		if error != nil {
			apiError.Code = http.StatusInternalServerError
			apiError.Message = error.Error()
		} else {
			if len(member.ID) > 0 {
				selected[member.ID] = member
				status = http.StatusOK
			} else {
				status = http.StatusNotFound
				apiError.Code = status
				apiError.Message = fmt.Sprintf("No member with ID %s", vars["id"])
			}
		}
	} else {
		status = http.StatusBadRequest
		apiError.Code = status
		apiError.Message = fmt.Sprintf("Missing id element in payload")
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

func (api API) DeleteMember(writer http.ResponseWriter, reader *http.Request) {
	status := http.StatusInternalServerError
	selected := make(map[string]storage.Member)
	vars := mux.Vars(reader)
	apiError := Error{}
	if id, ok := vars["id"]; ok {
		member, err := storage.Get(api.BoltDB, id)
		if err != nil {
			status = http.StatusInternalServerError
			apiError.Code = status
			apiError.Message = err.Error()
		} else if len(member.ID) == 0 {
			status = http.StatusNotFound
			apiError.Code = status
			apiError.Message = fmt.Sprintf("No member with ID %s", id)
		} else {
			err := storage.Delete(api.BoltDB, id)
			if err != nil {
				status = http.StatusInternalServerError
				apiError.Code = status
				apiError.Message = err.Error()
			}
			status = http.StatusNoContent
			return
		}
	} else {
		status = http.StatusBadRequest
		apiError.Code = status
		apiError.Message = fmt.Sprintf("Missing id field in request")
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

func (api API) UpdateMember(writer http.ResponseWriter, reader *http.Request) {
	status := http.StatusInternalServerError
	members := make(map[string]storage.Member)
	vars := mux.Vars(reader)
	apiError := Error{}
	if id, ok := vars["id"]; ok {
		exists, err := storage.CheckID(api.BoltDB, id)
		if err != nil {
			status = http.StatusInternalServerError
			apiError.Message = err.Error()
			apiError.Code = status
		} else {
			if exists {
				member := storage.Member{}
				err := json.NewDecoder(reader.Body).Decode(&member)
				if err == nil {
					member.ID = id // We ignore and replace whatever id they might have passed with the payload
					status = http.StatusOK
					members[id] = member
				}
			} else {
				status = http.StatusNotFound
				apiError.Code = status
				apiError.Message = fmt.Sprintf("No member with ID %s", id)
			}
		}
	} else {
		status = http.StatusBadRequest
		apiError.Code = status
		apiError.Message = fmt.Sprintf("Missing id field in request")
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

func (api API) NewMember(writer http.ResponseWriter, request *http.Request) {
	status := http.StatusInternalServerError
	var newMembers = make(map[string]storage.Member)
	var payload []byte
	var apiError = Error{}
	uuid, err := uuidGenerator.NewV4()
	if err != nil {
		status = http.StatusInternalServerError // Can change this to something else if more appropriate
		apiError.Code = status
		apiError.Message = "failed to generate UUID for member"
	} else {
		id := uuid.String()
		member := storage.Member{}
		err := json.NewDecoder(request.Body).Decode(&member)
		if err == nil {
			member.ID = id // We ignore whatever ID was specified in the POST request
			storage.Update(api.BoltDB, member)
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
