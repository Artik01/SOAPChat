package main

import (
	"log"
	"fmt"
	"os"
	"bufio"
	"encoding/xml"
	"github.com/gorilla/websocket"
)

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

func main() {
	var nickname string
	fmt.Print("Insert your nickname to join chat: ")
	fmt.Scan(&nickname)
	
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/socket", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	
	var data Data
	_, message, err := c.ReadMessage()
	if err != nil {
		log.Println("read:", err)
		return
	}
	err=xml.Unmarshal(message, &data)
	
	var room int
	for room <= 0 || room > data.Room {
		fmt.Printf("Choose room [1;%d]: ", data.Room)
		fmt.Scan(&room)
	}
	
	data.Nickname = nickname
	data.Room = room
	data.Msg = " joined"
	
	message, err =xml.Marshal(SOAPData(data))
	if err != nil {
		log.Println(err)
		return
	}
	
	err = c.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Println(err)
		return
	}
	
	reader := bufio.NewReader(os.Stdin)
	reader.ReadLine()//read '/n' that remain from fmt.Scan
	go Getter(c)
	for {
		msg, _, err := reader.ReadLine()
		if err != nil {
			log.Println("scan:", err)
		}
		data.Msg = ": "+string(msg)
		message, err =xml.Marshal(SOAPData(data))
		if err != nil {
			log.Println(err)
			return
		}
		
		err = c.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("read:", err)
			return
		}
	}
}

func Getter(c *websocket.Conn) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		var data Data
		xml.Unmarshal(message, &data)
		log.Printf("%s", data.Nickname+data.Msg)
	}
}

