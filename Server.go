
package main

import (
	"net/http"
	"io"
	"log"
	"github.com/gorilla/websocket"
	"encoding/xml"
)

var upgrader = websocket.Upgrader{}

var Rooms [][](*websocket.Conn)

var Mutexs [](chan int)

type SOAPData struct {
	XMLName xml.Name	`xml:"SOAP-ENV:Envelope"`
	Room int			`xml:"SOAP-ENV:Body>room"`
	Nickname string		`xml:"SOAP-ENV:Body>nick"`
	Msg string			`xml:"SOAP-ENV:Body>msg"`
}

type Data struct {
	XMLName xml.Name	`xml:"Envelope"`
	Room int			`xml:"Body>room"`
	Nickname string		`xml:"Body>nick"`
	Msg string			`xml:"Body>msg"`
}

func Handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Header().Set("Access-Control-Allow-Methods","POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "content-type")
	
	if req.Method == "POST" {
		data, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {return }
		
		log.Printf("%s\n", data)
		io.WriteString(w, "successful post")
	} else if req.Method == "OPTIONS" {
		w.WriteHeader(204)
	} else {
		w.WriteHeader(405)
	}
	
}

func Socket(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	var data Data
	data.Room = len(Rooms)
	p, err :=xml.Marshal(SOAPData(data))
	if err != nil {
		log.Println(err)
	}
	
	err = conn.WriteMessage(websocket.TextMessage, p)
	
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
	}
	
	
	xml.Unmarshal(p, &data)
	
	Rooms[data.Room-1] = append(Rooms[data.Room-1],conn)	
	
	SendMessageToOthers(data.Room, conn, messageType, p)
	
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			DeleteConn(data.Room, conn)
			conn.Close()
			data.Msg=" left"
			p, err =xml.Marshal(SOAPData(data))
			if err != nil {
				log.Println(err)
			}
			SendMessageToOthers(data.Room, nil, websocket.TextMessage, p)
			return
		}
	
		go SendMessageToOthers(data.Room, conn, messageType, p)
	}
}

func DeleteConn(room int, ToDel *websocket.Conn) {
	<- Mutexs[room-1]
	for i, conn := range Rooms[room-1] {
		if conn == ToDel {
			Rooms[room-1] = append(Rooms[room-1][:i], Rooms[room-1][i+1:]...)
			break
		}
	}
	Mutexs[room-1] <- 1
}

func SendMessageToOthers(room int, except *websocket.Conn, messageType int, p []byte) {
	for _, conn := range Rooms[room-1] {
		if conn == except {
			continue
		}
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	RoomsCnt := 5
	for i:=0; i < RoomsCnt; i++ {
		Mutexs = append(Mutexs, make(chan int, 1))
		Mutexs[i] <- 1
		Rooms = append(Rooms, nil)
	}
	http.HandleFunc("/", Handler)
	http.HandleFunc("/socket", Socket)
	
	err := http.ListenAndServe(":8080", nil)
	panic(err)
}
