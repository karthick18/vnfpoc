package vnfmgr
import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
)

type VnfRequest struct {
	Name string `json:"name"`
	Args string `json:"args"`
}

type VnfResponse struct {
	Id string `json:"id"`
	Name string `json:"name"`
}

type VnfResponses []VnfResponse

const (
	REST_API_ADDRESS = ":8081"
)

var (
	vnfMgr *VnfMgr = NewVnfMgr()
)

func StartREST() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/vnf/{id}", handleVnf).Methods("GET", "POST", "DELETE")
	router.HandleFunc("/vnf/update/{id}", updateVnf).Methods("POST")
	router.HandleFunc("/vnfs", listVnfs).Methods("GET")
	log.Fatal(http.ListenAndServe(REST_API_ADDRESS, router))
}

func handleVnf(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	id := vars["id"]
	switch req.Method {
	case "GET":
		vnf, ok := vnfMgr.Get(id)
		if !ok {
			res.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(res, "VNF %s not found\n", id)
			return
		}
		vnf_response := VnfResponse{Id:vnf.name, Name: vnf.name}
		vnf_json, err := json.Marshal(&vnf_response)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, string(vnf_json))
	case "POST":
		var vnfRequest VnfRequest
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&vnfRequest); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		future := vnfMgr.Dispatch(vnfRequest.Name, VNF_ADMIN_CREATE, vnfRequest.Args)
		if err := future.Get(); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(res, "%s\n", err.Error())
			return
		}
		res.WriteHeader(http.StatusCreated)
		fmt.Fprintf(res, "VNF %s created successfully\n", vnfRequest.Name)
	case "DELETE":
		future := vnfMgr.Dispatch(id, VNF_ADMIN_DELETE, "delete")
		err := future.Get()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(res, "%s\n", err.Error())
			return
		}
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "VNF %s deleted successfully\n", id)
	}
}

func updateVnf(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	//id is unused for now
	//vars := mux.Vars(req)
	//id := vars["id"]
	vnfRequest := new(VnfRequest)
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&vnfRequest)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	future := vnfMgr.Dispatch(vnfRequest.Name, VNF_ADMIN_UPDATE, vnfRequest.Args)
	err = future.Get()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(res, "%s\n", err.Error())
		return
	}
	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, "VNF %s updated successfully\n", vnfRequest.Name)
}

func listVnfs(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	var vnfResponses VnfResponses
	vnfs := vnfMgr.GetVnfs()
	for _, v := range vnfs {
		vnfResponse := VnfResponse{Id:v.name, Name:v.name}
		vnfResponses = append(vnfResponses, vnfResponse)
	}
	resp_json, err := json.Marshal(vnfResponses)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	fmt.Fprintf(res, string(resp_json))
}
