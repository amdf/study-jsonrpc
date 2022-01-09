package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ArgumentStruct struct {
	Number int
	Text   string
}

type ResultStruct struct {
	Result int
	Error  string
}

type InputData struct {
	ID     int          `json:"id"`
	Result ResultStruct `json:"result"`
	Err    error        `json:"error"`
}

type OutputData struct {
	JSONRPC string
	ID      int
	Method  string
	Params  []ArgumentStruct
}

type SomeServiceI interface {
	SomeMethod(in *ArgumentStruct, out *ResultStruct) error
}

type SomeService struct {
	SomeServiceI
	client *http.Client
}

func (s *SomeService) reqForSomeMethod(in *ArgumentStruct) (req *http.Request, err error) {
	arg := OutputData{
		JSONRPC: "2.0",
		ID:      22222,
		Method:  "SomeService.SomeMethod",
		Params: []ArgumentStruct{
			*in,
		},
	}
	var data []byte

	data, err = json.Marshal(arg)
	if err != nil {
		return
	}

	reader := bytes.NewReader(data)

	req, err = http.NewRequest("POST", "http://localhost:8081/rpc", reader)
	return
}

func (s *SomeService) SomeMethod(in *ArgumentStruct, out *ResultStruct) (err error) {
	var req *http.Request
	var resp *http.Response
	req, err = s.reqForSomeMethod(in)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if nil == err {
		var indata InputData
		err = json.Unmarshal(data, &indata)
		if nil == err {
			*out = indata.Result
		} else {
			err = errors.New("not ok")
		}
	}
	return err
}

func NewService() *SomeService {
	return &SomeService{client: &http.Client{}}
}

func main() {
	svc := NewService()
	var res ResultStruct

	err := svc.SomeMethod(&ArgumentStruct{Number: 1, Text: "test"}, &res)
	if nil == err {
		fmt.Println("SomeMethod called!")
		fmt.Println("Result:", res)
	} else {
		fmt.Println(err)
	}
}
