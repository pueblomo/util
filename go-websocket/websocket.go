package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"strconv"

	"github.com/gorilla/websocket"
)

type MessageWrapper struct {
	WrapperType string `json:"type"`
	Data        string `json:"data"`
}

type Message struct {
	MessageType string `json:"type"`
	FileName    string `json:"fileName"`
	Data        string `json:"data"`
	Number      int    `json:"number"`
	IsLast      bool   `json:"isLast"`
}

func main() {
	// Parse the WebSocket server URL.
	serverURL, err := url.Parse("ws://localhost:8080/ws")
	if err != nil {
		log.Fatal(err)
	}

	// Create a new WebSocket connection.
	conn, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Start a goroutine to read messages from the WebSocket connection.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	go func() {
		wrapper, err := json.Marshal(MessageWrapper{"INITIAL", base64.StdEncoding.EncodeToString([]byte("Mein Go Client " + strconv.Itoa(rand.Intn(100))))})
		if err != nil {
			log.Println("marshal:", err)
			return
		}

		log.Println("Send initial message")

		err = conn.WriteMessage(websocket.TextMessage, []byte(wrapper))
		if err != nil {
			log.Println("write:", err)
			return
		}
	}()

	go func() {
		fileContent, err := ioutil.ReadFile("test.txt")
		if err != nil {
			log.Println("Error reading file:", err)
			return
		}

		halfSize := len(fileContent) / 2
		contentOne := fileContent[:halfSize]
		contentTwo := fileContent[halfSize:]

		fileName := "test" + strconv.Itoa(rand.Intn(100)) + ".txt"

		m1, err := json.Marshal(Message{"CREATE", fileName, base64.StdEncoding.EncodeToString(contentOne), 1, false})
		m2, err := json.Marshal(Message{"CREATE", fileName, base64.StdEncoding.EncodeToString(contentTwo), 2, true})

		wrapperOne, err := json.Marshal(MessageWrapper{"FILE", base64.StdEncoding.EncodeToString(m1)})
		wrapperTwo, err := json.Marshal(MessageWrapper{"FILE", base64.StdEncoding.EncodeToString(m2)})

		log.Println("Send message one")
		err = conn.WriteMessage(websocket.TextMessage, []byte(wrapperOne))
		if err != nil {
			log.Println("write:", err)
			return
		}

		log.Println("Send message two")
		err = conn.WriteMessage(websocket.TextMessage, []byte(wrapperTwo))
		if err != nil {
			log.Println("write:", err)
			return
		}

	}()

	// Wait for a signal to interrupt the program.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Wait for either a signal or for the goroutines to finish.
	select {
	case <-interrupt:
		log.Println("interrupt")
	case <-done:
		log.Println("done")
	}
}
