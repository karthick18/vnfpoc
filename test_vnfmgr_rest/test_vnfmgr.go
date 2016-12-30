package main
import (
	"fmt"
	"github.com/karthick18/vnfpoc/vnfmgr"
	"net/http"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	VNFMGR_BASE_REQUEST_URL = "http://127.0.0.1:8081/vnf/"
	VNFMGR_BASE_UPDATE_URL = "http://127.0.0.1:8081/vnf/update/"
	VNFMGR_LIST_REQUEST_URL = "http://127.0.0.1:8081/vnfs"
	HTTP_HEADER = "application/json"
)

type HttpResponse struct {
	status string
	body []byte
	err error
}

func create_vnfs(numVnfs int) {
	var channels []chan HttpResponse
	for i := 0; i < numVnfs; i++ {
		channel := make(chan HttpResponse)
		channels = append(channels, channel)
	}
	for i := 0; i < numVnfs; i++ {
		name := fmt.Sprintf("vnf_%d", i)
		args := fmt.Sprintf("args_%d", i)
		vnfRequest := vnfmgr.VnfRequest{Name:name, Args: args}
		go func(vnfRequest vnfmgr.VnfRequest, c chan HttpResponse) {
			url := VNFMGR_BASE_REQUEST_URL + vnfRequest.Name
			request, err := json.Marshal(&vnfRequest)
			if err != nil {
				panic(err)
			}
			resp, err := http.Post(url, HTTP_HEADER, bytes.NewBuffer(request))
			if err != nil {
				httpResponse := HttpResponse{body:[]byte{}, err: err}
				c <- httpResponse
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				httpResponse := HttpResponse{status: resp.Status, body:[]byte{}, err: err}
				c <- httpResponse
				return
			}
			httpResponse := HttpResponse{status: resp.Status, body: body, err: err}
			c <- httpResponse
		} (vnfRequest, channels[i])
	}
	for i := 0; i < numVnfs; i++ {
		httpResponse := <- channels[i]
		if httpResponse.err != nil {
			fmt.Println(httpResponse.err)
			continue
		}
		fmt.Println("Got response", string(httpResponse.body))
	}
}

func update_vnfs(numVnfs int) {
	var channels []chan HttpResponse
	for i := 0; i < numVnfs; i++ {
		channel := make(chan HttpResponse)
		channels = append(channels, channel)
	}
	for i := 0; i < numVnfs; i++ {
		name := fmt.Sprintf("vnf_%d", i)
		args := fmt.Sprintf("args_%d", i)
		vnfRequest := vnfmgr.VnfRequest{Name:name, Args: args}
		go func(vnfRequest vnfmgr.VnfRequest, c chan HttpResponse) {
			url := VNFMGR_BASE_UPDATE_URL + vnfRequest.Name
			request, err := json.Marshal(&vnfRequest)
			if err != nil {
				panic(err)
			}
			resp, err := http.Post(url, HTTP_HEADER, bytes.NewBuffer(request))
			if err != nil {
				httpResponse := HttpResponse{body:[]byte{}, err: err}
				c <- httpResponse
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				httpResponse := HttpResponse{status: resp.Status, body:[]byte{}, err: err}
				c <- httpResponse
				return
			}
			httpResponse := HttpResponse{status: resp.Status, body: body, err: err}
			c <- httpResponse
		} (vnfRequest, channels[i])
	}
	for i := 0; i < numVnfs; i++ {
		httpResponse := <- channels[i]
		if httpResponse.err != nil {
			fmt.Println(httpResponse.err)
			continue
		}
		fmt.Println("Got response", string(httpResponse.body))
	}
}

func list_vnfs() {
	url := VNFMGR_LIST_REQUEST_URL
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var vnfResponses vnfmgr.VnfResponses
	if err := json.Unmarshal(body, &vnfResponses); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("List request returned")
	for _, vnf_resp := range vnfResponses {
		fmt.Println(vnf_resp.Id, vnf_resp.Name)
	}
	fmt.Println("")
}

func delete_vnfs(numVnfs int) {
	var channels []chan HttpResponse
	for i := 0; i < numVnfs; i++ {
		channel := make(chan HttpResponse)
		channels = append(channels, channel)
	}
	for i := 0; i < numVnfs; i++ {
		name := fmt.Sprintf("vnf_%d", i)
		go func(id string, c chan HttpResponse) {
			url := VNFMGR_BASE_REQUEST_URL + id
			req, err := http.NewRequest("DELETE", url, nil)
			if err != nil {
				panic(err)
			}
			resp, err := http.DefaultClient.Do(req)
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				httpResponse := HttpResponse{status: resp.Status, body:[]byte{}, err: err}
				c <- httpResponse
				return
			}
			httpResponse := HttpResponse{status: resp.Status, body: body, err: err}
			c <- httpResponse
		} (name, channels[i])
	}
	for i := 0; i < numVnfs; i++ {
		httpResponse := <- channels[i]
		if httpResponse.err != nil {
			fmt.Println(httpResponse.err)
			continue
		}
		fmt.Println("Got response", string(httpResponse.body))
	}
}

func main() {
	var numVnfs int = 10
	if len(os.Args) > 1 {
		numVnfs, _ = strconv.Atoi(os.Args[1])
		if numVnfs <= 0 {
			numVnfs = 10
		}
	}
	fmt.Println("Starting vnfmgr REST api server")
	go vnfmgr.StartREST()
	create_vnfs(numVnfs)
	update_vnfs(numVnfs)
	list_vnfs()
	delete_vnfs(numVnfs)
	update_vnfs(numVnfs)
}
