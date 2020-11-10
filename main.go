package main

import (
	"github.com/tarm/serial"
	"net/http"
	"log"
	"io/ioutil"
	"github.com/gorilla/websocket"
	"fmt"
	"strings"
	// "time"
)
var (
	c = &serial.Config{Name: "/dev/tty.MANPOWER_THERMOGUN-ESP3", Baud: 115200}
	s,err = serial.OpenPort(c)
	dataRead = "0"
	conn *websocket.Conn
)

type msg struct {
	Num int
}

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func main() {
	go read()
	// http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "index.html")
    })
	http.HandleFunc("/data", dataHandler)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _ = upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
		
        for {
			// Read message from browser
			go readSocket(conn)
            msgType, msg, err := conn.ReadMessage()
            if err != nil {
                return
            }

            // Print the message to the console
            fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

            // Write message back to browser
            if err = conn.WriteMessage(msgType, msg); err != nil {
                return
            }
        }
    })

	
	panic(http.ListenAndServe(":8181", nil))
	
}

func read() {
	buf := make([]byte, 1024)
	for {
		n, err := s.Read(buf)
		if err != nil {
				log.Fatal(err)
		}
		data := string(buf[:n])
		if len(strings.TrimSpace(data)) != 0 {
			dataRead = data
			fmt.Println(data)
			// fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), data)
			// conn, _ =  upgrader.()
			
		}
	}
	// read()
}

func readSocket(conn *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := s.Read(buf)
		if err != nil {
				log.Fatal(err)
		}
		data := string(buf[:n])
		if len(strings.TrimSpace(data)) != 0 {
			dataRead = data
			fmt.Println(data)
			// fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), data)
			// conn, _ =  upgrader.()
			if err = conn.WriteMessage(n, buf); err != nil {
                // return
            }
			
		}
	}
	// read()
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("index.html")
	if err != nil {
		fmt.Println("Could not open file.", err)
	}
	fmt.Fprintf(w, "%s", content)
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s", dataRead)
}


func wsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("User Connected.")
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	go echo(conn)
}

func echo(conn *websocket.Conn) {
	for {
		m := msg{}

		err := conn.ReadJSON(&m)
		if err != nil {
			fmt.Println("Error reading json.", err)
		}

		fmt.Printf("Got message: %#v\n", m)

		if err = conn.WriteJSON(m); err != nil {
			fmt.Println(err)
		}
	}
}