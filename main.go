package main // import "github.com/h0ru5/sonoff-lanmode-switch"

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func getInitMsg() ([]byte, error) {
	initMsg := map[string]interface{}{
		"action":    "userOnline",
		"ts":        time.Now().Unix(),
		"version":   6,
		"apikey":    "nonce",
		"sequence":  time.Now(),
		"userAgent": "app",
	}
	return json.Marshal(initMsg)
}

func getUpdateMessage(newState *string) ([]byte, error) {
	updateMsg := map[string]interface{}{
		"action":     "update",
		"deviceid":   "nonce",
		"apikey":     "nonce",
		"selfApikey": "nonce",
		"params": map[string]interface{}{
			"switch": *newState,
		},
		"sequence":  time.Now(),
		"userAgent": "app",
	}
	return json.Marshal(updateMsg)
}

func updateState(device *string, newState *string) {
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:8081", *device),
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s\n", message)
		}
	}()

	initMsg, err := getInitMsg()
	if err != nil {
		log.Fatal("Marshalling:", err)
		return
	}
	c.WriteMessage(websocket.TextMessage, initMsg)
	fmt.Println("sent", string(initMsg))

	updateMsg, err := getUpdateMessage(newState)
	if err != nil {
		log.Fatal("Marshalling:", err)
		return
	}
	c.WriteMessage(websocket.TextMessage, updateMsg)
	fmt.Println("sent", string(updateMsg))

	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
}

func main() {
	device := flag.String("ip", "192.168.178.42", "IP address of sonoff device")
	off := flag.Bool("off", false, "overrides default (on) to switch off")
	flag.Parse()

	newState := "on"
	if *off {
		newState = "off"
	}

	fmt.Printf("Attempting to switch device %s to %s\n", *device, newState)

	updateState(device, &newState)

}
