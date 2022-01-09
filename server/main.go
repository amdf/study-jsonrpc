package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

// adapt HTTP connection to ReadWriteCloser
type HttpConn struct {
	in  io.Reader
	out io.Writer
}

func (c *HttpConn) Read(p []byte) (n int, err error)  { return c.in.Read(p) }
func (c *HttpConn) Write(d []byte) (n int, err error) { return c.out.Write(d) }
func (c *HttpConn) Close() error                      { return nil }

type Handler struct {
	rpcServer *rpc.Server
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serverCodec := jsonrpc.NewServerCodec(&HttpConn{
		in:  r.Body,
		out: w,
	})
	w.Header().Set("Content-type", "application/json")
	err := h.rpcServer.ServeRequest(serverCodec)
	if err != nil {
		http.Error(w, `{"error":"cant serve request"}`, 500)
	}
}

type ArgumentStruct struct {
	Number int
	Text   string
}

type ResultStruct struct {
	Result int
	Error  string
}

type SomeServiceI interface {
	SomeMethod(in *ArgumentStruct, out *ResultStruct) error
}

type SomeService struct {
	SomeServiceI
}

func (s *SomeService) SomeMethod(in *ArgumentStruct, out *ResultStruct) error {
	fmt.Println("SomeMethod called!")
	fmt.Println("Params:", in)
	out.Result = 12345
	if in.Number != 0 {
		out.Error = "Error: " + in.Text
	}
	return nil
}

func NewService() *SomeService {
	return &SomeService{}
}

func main() {
	svc := NewService()
	server := rpc.NewServer()
	err := server.Register(svc)
	if err != nil {
		log.Fatal("RPC register error:", err)
	}
	svcHandler := &Handler{
		rpcServer: server,
	}
	http.Handle("/rpc", svcHandler)
	fmt.Println("start...")
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("RPC serve error:", err)
	}
}
