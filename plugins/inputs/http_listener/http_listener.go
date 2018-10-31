package http_listener

import (
	"crypto/subtle"
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/plugins/parsers"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	DEFAULT_MAX_BODY_SIZE = 5 * 1024 * 1024 // 500 MB
	DEFAULT_MAX_LINE_SIZE = 64 * 1024       // 64 KB
)

type HTTPListener struct {
	Address       string
	parsers       map[string]parsers.Parser
	NameOverride  string
	MaxBodySize   int64
	MaxLineSize   int
	ReadTimeout   internal.Duration
	WriteTimeout  internal.Duration
	BasicUsername string
	BasicPassword string

	listener net.Listener
	mu       sync.Mutex
	wg       sync.WaitGroup
	acc      deviceAgent.Accumulator

	contentMap map[string][]byte
}

func (h *HTTPListener) SetParser(parsers map[string]parsers.Parser) {
	for k, v := range parsers {
		h.parsers[k] = v
	}
}

func (h *HTTPListener) SelfCheck() deviceAgent.Quality {
	return deviceAgent.QualityGood
}

func (h *HTTPListener) Name() string {
	return "http_listener"
}

func (h *HTTPListener) Gather(deviceAgent.Accumulator) error {
	//for k, v := range h.parsers {
	//	log.Println(k, v)
	//}
	return nil
}

func (h *HTTPListener) SetPointMap(map[string]deviceAgent.PointDefine) {

}

func renderMsg(w http.ResponseWriter, message string, statusCode ...int) {
	w.Header().Set("Content-Type", "application/json")
	if len(statusCode) > 0 {
		w.WriteHeader(statusCode[0])
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	b, _ := json.Marshal(map[string]interface{}{"msg": message})
	w.Write(b)
}

func (h *HTTPListener) serveWrite(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > h.MaxBodySize {
		renderMsg(w, "http: request body too large",
			http.StatusRequestEntityTooLarge)
		return
	}
}
func (h *HTTPListener) serveFile(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > h.MaxBodySize {
		renderMsg(w, "http: request body too large",
			http.StatusRequestEntityTooLarge)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.MaxBodySize)
	if err := r.ParseMultipartForm(h.MaxBodySize); err != nil {
		log.Println("E! " + err.Error())
		renderMsg(w, err.Error())
		return
	}

	for k := range r.MultipartForm.File {
		file, _, err := r.FormFile(k)
		if err != nil {
			renderMsg(w, err.Error())
			return
		}
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderMsg(w, err.Error())
			return
		}
		fileType := http.DetectContentType(fileBytes)
		if fileType != "text/plain; charset=utf-8" {
			renderMsg(w, "only support utf-8 text file now")
			return
		}
		h.contentMap[k] = fileBytes
		file.Close()
	}
	renderMsg(w, "OK", 200)
}

func (h *HTTPListener) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.MaxBodySize == 0 {
		h.MaxBodySize = DEFAULT_MAX_BODY_SIZE
	}
	if h.MaxLineSize == 0 {
		h.MaxLineSize = DEFAULT_MAX_LINE_SIZE
	}
	if h.ReadTimeout.Duration < time.Second {
		h.ReadTimeout.Duration = time.Second * 10
	}
	if h.WriteTimeout.Duration < time.Second {
		h.WriteTimeout.Duration = time.Second * 10
	}

	server := &http.Server{
		Addr:         h.Address,
		Handler:      h,
		ReadTimeout:  h.ReadTimeout.Duration,
		WriteTimeout: h.WriteTimeout.Duration,
	}
	listener, err := net.Listen("tcp", h.Address)
	if err != nil {
		return err
	}
	h.listener = listener
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		server.Serve(h.listener)
	}()
	log.Printf("I! Started HTTP listener service on %s\n", h.Address)

	return nil
}
func (h *HTTPListener) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	urlPathList := strings.Split(req.URL.Path, "/")
	switch urlPathList[1] {
	case "file":
		h.AuthenticateIfSet(h.serveFile, res, req)
	case "write":
		h.AuthenticateIfSet(h.serveWrite, res, req)
	case "ping":
		h.AuthenticateIfSet(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusNoContent)
		}, res, req)
	default:
		h.AuthenticateIfSet(http.NotFound, res, req)
	}
}

func (h *HTTPListener) AuthenticateIfSet(handler http.HandlerFunc, res http.ResponseWriter, req *http.Request) {
	if h.BasicUsername != "" && h.BasicPassword != "" {
		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(reqUsername), []byte(h.BasicUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPassword), []byte(h.BasicPassword)) != 1 {

			http.Error(res, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		handler(res, req)
	} else {
		handler(res, req)
	}
}

func (h *HTTPListener) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listener.Close()
	h.wg.Wait()
	log.Println("I! Stopped HTTP listener service on ", h.Address)
}

func init() {
	inputs.Add("http_listener", func() deviceAgent.Input {
		return &HTTPListener{
			parsers:    make(map[string]parsers.Parser),
			contentMap: make(map[string][]byte),
		}
	})
}
