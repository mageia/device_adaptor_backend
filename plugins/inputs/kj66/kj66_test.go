package kj66

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

import (
	"log"
)

const (
	server = "192.168.127.135:4001"
)

var quitControl = make(chan bool)
var switchStatus = make([]byte, 8)

var cmdRunning = false

type cmdError struct {
	msg string
}

func (ce cmdError) Error() string {
	return ce.msg
}

type TcpClient struct {
	connection *net.TCPConn
	hawkServer *net.TCPAddr
	stopChan   chan struct{}
}

func crc16Sick(buf []byte) uint16 {
	var wCRC uint16
	var preByte uint8

	for i := 0; i < len(buf); i++ {
		if wCRC&0x8000 != 0 {
			wCRC = (wCRC << 1) ^ 0x8005
		} else {
			wCRC <<= 1
		}

		wCRC ^= uint16(preByte)*256 + uint16(buf[i])
		preByte = buf[i]
	}

	return wCRC
}

func (c *TcpClient) SendCmd(quit chan bool, t string) ([]byte, error) {
	cmd := []byte{0x5a, 0xff, 0x06, 'B', 0x01, 0x01, 0, 0, 0}

	switch t {
	case "open":
		cmd[5] = 1
		fmt.Println("Open cmd")
	case "close":
		cmd[5] = 0
		fmt.Println("Close cmd")
	default:
		fmt.Println("Unknown cmd")
		return nil, cmdError{"Unknown cmd"}
	}

	crc := crc16Sick(cmd[1:7])
	cmd[7], cmd[8] = byte(crc>>8), byte(crc&0xff)
	fmt.Printf("Write cmd: %v\n", cmd)

	for {
		n, e := c.connection.Write(cmd)
		if e != nil {
			return nil, cmdError{"Write cmd Failed"}
		}
		fmt.Printf("Cmd exec success, write [%d] bytes\n", n)

		rsp := c.GetResponse()
		switchStatus = rsp
		fmt.Printf("Read response: %v\n", rsp)

		select {
		case <-quit:
			log.Printf("End cmd: %s", t)
			return rsp, nil
		case now := <-time.After(1 * time.Second):
			fmt.Println(now)
		}
	}
}

func (c *TcpClient) GetResponse() []byte {
	reader := bufio.NewReader(c.connection)

	buf := make([]byte, 20)
	n, e := reader.Read(buf)
	if e != nil {
		fmt.Printf("End read, n: [%d], err: [%v]", n, e)
	}

	return buf[:n]
}

func SimpleHTTPServer(c *TcpClient) {
	http.HandleFunc("/control", func(w http.ResponseWriter, r *http.Request) {
		if cmdRunning {
			quitControl <- true
		}
		go c.SendCmd(quitControl, r.URL.Query().Get("cmd"))
		cmdRunning = true

		fmt.Fprint(w, "Cmd exec success\n")
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		type StatusResponse struct {
			KzStatus byte `json:"kz_status"`
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode(StatusResponse{switchStatus[4]})
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func main() {
	hawkServer, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		fmt.Printf("hawk server: %s error: [%s]", server, err.Error())

	}
	conn, err := net.DialTCP("tcp", nil, hawkServer)
	if err != nil {
		fmt.Printf("connect to server: %s error: [%s]", server, err.Error())
	}
	client := &TcpClient{
		connection: conn,
		hawkServer: hawkServer,
		stopChan:   make(chan struct{}),
	}

	go SimpleHTTPServer(client)

	<-client.stopChan

	client.connection.Close()
}
