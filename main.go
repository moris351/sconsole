package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	zmq "github.com/pebbe/zmq4"
	_ "time"
)

var addr = flag.String("addr", ":8080", "http service address")

const (
	receiverPort = ":5557"
	senderPort   = ":5558"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type zmqTool struct {
	receiver *zmq.Socket
	sender   *zmq.Socket
}

func reader(c *websocket.Conn, zt *zmqTool) {
	for {

		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		zt.sender.Send(string(message), 0)
	}
}
func writer(c *websocket.Conn, zt *zmqTool) {
	for {
		s, err := zt.receiver.Recv(0)
		if err != nil {
			log.Println("receiver.Recv err, ", err)
		}
		if len(s) > 0 {
			log.Println(s)
			err := c.WriteMessage(websocket.TextMessage, []byte(s))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
		//time.Sleep(time.Second)
	}

}

func echo(zt *zmqTool, w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	go writer(c, zt)
	reader(c, zt)
}
func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}
func newZmqTool() *zmqTool {
	zt := new(zmqTool)
	var err error
	zt.receiver, err = zmq.NewSocket(zmq.PULL)
	if err != nil {
		log.Println("zmq NewSocket err,", err)
		return nil
	}
	//defer sender.Close()
	zt.receiver.Bind("tcp://*" + receiverPort)

	zt.sender, err = zmq.NewSocket(zmq.PUSH)
	//defer receiver.Close()
	if err != nil {
		log.Println("zmq NewSocket err,", err)
		return nil
	}
	zt.sender.Bind("tcp://*" + senderPort)
	return zt
}

func (zt *zmqTool) close() {
	zt.receiver.Close()
	zt.sender.Close()
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	zt := newZmqTool()
	defer zt.close()

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		echo(zt, w, r)
	})

	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("server.html").ParseFiles("tmpls/server.html"))
