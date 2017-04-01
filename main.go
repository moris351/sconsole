package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"time"
	"github.com/gorilla/websocket"
	zmq "github.com/pebbe/zmq4"
)

var addr = flag.String("addr", ":8080", "http service address")

var receiver *zmq.Socket
var sender *zmq.Socket
//var receq chan string
//var sendq chan string

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
} 


func reader(c *websocket.Conn){
	for{
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
	}
}
func writer(c *websocket.Conn){
	for{
		s,err := receiver.Recv(0)
		if(err!=nil){
			log.Println("receiver.Recv err, ", err)
		}
		if(len(s)>0){
			log.Println(s)
			err := c.WriteMessage(websocket.TextMessage, []byte(s))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
		time.Sleep(time.Second)
	}

}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	go writer(c)
	reader(c)
}
func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	var err error
	sender,err = zmq.NewSocket(zmq.PUSH)
	if(err!=nil){
		log.Println("sender socket new err,", err)
	}
	defer sender.Close()
	sender.Bind("tcp://*:5558")

	receiver,err = zmq.NewSocket(zmq.PULL)
	if(err!=nil){
		log.Println("receiver socket new err,", err)
	}
	defer receiver.Close()

	receiver.Bind("tcp://*:5557")

	//receq = make(chan string)
	//sendq = make(chan string)

	http.HandleFunc("/echo", echo)

	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("server.html").ParseFiles("tmpls/server.html"))
