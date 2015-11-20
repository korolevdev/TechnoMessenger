package server

import (
	"../utils"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

const PORT = "7777"

type Client struct {
	uid      string
	login    string
	incoming chan string
	outgoing chan string
	reader   *bufio.Reader
	writer   *bufio.Writer
}

func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)
	client := &Client{
		uid:      "",
		login:    "",
		incoming: make(chan string),
		outgoing: make(chan string),
		reader:   reader,
		writer:   writer,
	}
	//client.Listen()
	return client
}

func main() {

	// 1 Connect
	conn, err := net.Dial("tcp", ":"+PORT)
	utils.CheckError(err, "Can't connect to server", true)

	client := NewClient(conn)

	defer conn.Close()

	//go
	watchForConnection(client)
}

func (c Client) Login() {
	s, err := json.Marshal(struct {
		Action string        `json:"action"`
		Data   utils.CltAuth `json:"data"`
	}{
		Action: "auth",
		Data:   utils.CltAuth{Login: "user", Pass: "12345"},
	})
	//fmt.Printf("Result %v\n", string(s))
	if utils.CheckError(err, "Login error", false) {
		c.writer.Write(s)
		c.writer.Flush()
	}
}

func (c Client) GetChannelList() {
}

func watchForConnection(client *Client) {
	dec := json.NewDecoder(client.reader)
	for {
		var m utils.SrvMessage
		err := dec.Decode(&m)
		utils.CheckError(err, "Invalid message", true)
		switch m.Action {
		case "welcome":
			fmt.Println("Welcome message")
			client.Login()
		case "auth", "register", "leave":
			fmt.Printf("Message %v\n", m.Action)
			var im utils.SrvStatusMessage
			if err := json.Unmarshal(m.RawData, &im); err != nil {
				log.Fatal(err)
				return
			}

			if im.Status == 0 {
				client.GetChannelList()
			} else {
				fmt.Printf("Auth error - %v", im.Error)
				client.Login()
			}

		case "channellist":
			fmt.Printf("channellist Message\n")
		default:
			fmt.Printf("Unknown Message %v\n", m.Action)
		}
	}
}
