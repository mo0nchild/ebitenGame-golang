package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	event "../../game"
	"github.com/gorilla/websocket"
)

type userDataBuffer struct {
	Data []struct{} `json:"server"`
}

var (
	dataBuffer userDataBuffer
	units      event.Units
	world      event.World
)

func init() {
	world = event.World{Units: event.Units{}}
}

func (udb userDataBuffer) sendDataToClient(conn *websocket.Conn, world event.World) error {
	data := event.Event{
		Type: event.EventTypeUpdate,
		Data: world.Units,
	}
	jsonData, _ := json.Marshal(data)
	if err := conn.WriteJSON(jsonData); err != nil {
		return err
	}
	return nil
}

func jsonDataPage(w http.ResponseWriter, r *http.Request) {

	bytes, err := json.Marshal(dataBuffer)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprintf(w, string(bytes))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func updateRequest(conn *websocket.Conn, world event.World) {
	var (
		byteData []byte
		data     event.Event
	)

	go func() {
		for {
			dataBuffer.sendDataToClient(conn, world)
			time.Sleep(1 * time.Second / 60)
		}
	}()
	for {
		if err := conn.ReadJSON(&byteData); err != nil {
			log.Println(err)
			return
		}

		json.Unmarshal(byteData, &data)
		world.EventHandler(data)

		file, _ := json.MarshalIndent(dataBuffer, "", "")
		ioutil.WriteFile("data.json", file, 0600)

	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client Connected")

	var addPlayer []byte
	if err := ws.ReadJSON(&addPlayer); err != nil {
		log.Println(err)
		return
	}

	var data event.Event
	json.Unmarshal(addPlayer, &data)

	world.EventHandler(data)

	updateRequest(ws, world)
}

func determineListenAddress() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return ":8080", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}

func searchIPAddress() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				log.Println(ipnet.IP.String())
			}
		}
	}
}

func main() {
	addr, err := determineListenAddress()
	if err != nil {
		log.Println(err)
	}
	log.Println(addr)
	searchIPAddress()

	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/data", jsonDataPage)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
