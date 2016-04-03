package server

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"time"
)

// Client class of client
type Client struct {
	conn   net.Conn // Connection to user socket
	uid    string   // UserID
	login  string   // Login of user
	ip     string   // Ip of client
	cid    string   // Client ID
	sid    string   // Session ID
	nick   string   // Nickname of user
	status string   // Stirng of user's status
	avatar string   // Picture of user
	email  string   // User's email
	phone  string   // User's phone

	connected       bool // Connection user state
	offlineMessages [][]byte

	outgoing chan []byte
	reader   *bufio.Reader
	writer   *bufio.Writer
	contacts map[string]string // Map of uids of users (key uid; value uid)
}

// write - send data to user
func (c *Client) write() {
	for data := range c.outgoing {
		if c.connected {
			c.writer.Write(data)
			c.writer.Flush()
		} else {
			c.offlineMessages = append(c.offlineMessages, data)
		}
	}
}

// Listen - start corotinues for listening and writing
func (c *Client) Listen() {
	go c.read()
	go c.write()
}

// NewClient create new instance of Client class
func NewClient(connection net.Conn) *Client {
	writer := bufio.NewWriter(connection)
	reader := bufio.NewReader(connection)
	client := &Client{
		conn:     connection,
		uid:      "",
		login:    "",
		status:   "",
		email:    "",
		phone:    "",
		ip:       connection.RemoteAddr().String(),
		outgoing: make(chan []byte),
		reader:   reader,
		writer:   writer,
		contacts: make(map[string]string),

		connected:       true,
		offlineMessages: make([][]byte, 0),
	}
	//client.Listen()
	return client
}

// Register register new user on server
func (c *Client) Register(login string, pass string, nick string) {
	status, err := gServer.Register(c, login, pass, nick)
	if err != nil {
		c.Error("register", err.Error(), status, true)
		return
	}
	s, err := json.Marshal(struct {
		Action string           `json:"action"`
		Data   SrvStatusMessage `json:"data"`
	}{
		Action: "register",
		Data: SrvStatusMessage{
			Status: ErrOK,
			Error:  "OK",
		},
	})
	if !c.CheckError(err, "Can't marhsal answer") {
		return
	}
	c.outgoing <- s

	c.Auth(login, pass)
}

// Disconnect is function a wrapper of conn.Close
func (c *Client) Disconnect() {
	c.conn.Close()
	c.connected = false
}

// CheckError wrapper to check err construction
func (c *Client) CheckError(err error, message string) bool {
	if err != nil {
		log.Println("(" + c.ip + ") " + message + ": " + err.Error())
	}
	return err == nil
}

// SetUserInfo is updates information about user
func (c *Client) SetUserInfo(ava string, email string, phone string, userstatus string) {

	c.avatar = ava
	c.status = userstatus

	gServer.UpdateUserData(c, email, phone)

	if email != "" {
		c.email = email
	}

	if phone != "" {
		c.phone = phone
	}

	c.Ok("setuserinfo")
}

// GetContactList sends contact list to user
func (c *Client) GetContactList() {
	list := SrvListOfUsers{}
	list.Status = ErrOK
	list.Error = "OK"
	list.Users = make([]UserData, 0)

	for _, uid := range c.contacts {
		if user, ok := gServer.GetUserData(uid); ok {
			list.Users = append(list.Users, UserData{
				Uid:    uid,
				Nick:   user.nick,
				Email:  user.email,
				Phone:  user.phone,
				Avatar: user.avatar,
			})
		}
	}

	m, err := json.Marshal(struct {
		Action string         `json:"action"`
		Data   SrvListOfUsers `json:"data"`
	}{
		Action: "contactlist",
		Data:   list,
	})

	if !c.CheckError(err, "Can't marhsal answer") {
		return
	}
	c.outgoing <- m
}

// ImportContacts finds users contacts on server
func (c *Client) ImportContacts(contacts []Contact) {
	list := SrvListOfUsers{}
	list.Status = ErrOK
	list.Error = "OK"
	list.Users = make([]UserData, 0)

	for _, contact := range contacts {
		if user, ok := gServer.FindUser(contact.Email, contact.Phone); ok {
			list.Users = append(list.Users, UserData{
				Uid:    user.login,
				Nick:   user.nick,
				Email:  user.email,
				Phone:  user.phone,
				Avatar: user.avatar,
				MyID:   contact.MyID,
			})
		}
	}

	m, err := json.Marshal(struct {
		Action string         `json:"action"`
		Data   SrvListOfUsers `json:"data"`
	}{
		Action: "import",
		Data:   list,
	})

	if !c.CheckError(err, "Can't marhsal answer") {
		return
	}
	c.outgoing <- m
}

// AddContact adds contact to user list
func (c *Client) AddContact(uid string) {
	_, ok := c.contacts[uid]

	// Already has contact
	if ok || uid == c.uid {
		c.Error("addcontact", "User already in list", ErrAlreadyExist, true)
		return
	}

	_, ok = gServer.GetUserData(uid)
	if !ok {
		c.Error("addcontact", "User not found", ErrUserNotFound, true)
		return
	}
	c.contacts[uid] = uid

	c.Ok("addcontact")
}

// DelContact removes contact from user list
func (c *Client) DelContact(uid string) {
	delete(c.contacts, uid)
	c.Ok("delcontact")
}

// Auth client autorisation on server
func (c *Client) Auth(login string, pass string) bool {
	sid, status, err := gServer.Auth(c, login, pass)
	if err != nil {
		c.Error("auth", err.Error(), status, true)
		return false
	}
	c.login = login
	c.uid = login
	c.sid = sid
	m := SrvStatusAuthMessage{
		Sid:  c.sid,
		Cid:  c.cid,
		Nick: c.nick,
	}
	m.Status = ErrOK
	m.Error = "OK"
	s, err := json.Marshal(struct {
		Action string               `json:"action"`
		Data   SrvStatusAuthMessage `json:"data"`
	}{
		Action: "auth",
		Data:   m,
	})
	if !c.CheckError(err, "Can't marhsal message") {
		c.Disconnect()
		return false
	}
	c.outgoing <- s

	// Send offlineMessages
	for _, mess := range c.offlineMessages {
		c.outgoing <- mess
	}
	c.offlineMessages = make([][]byte, 0)

	return true
}

// Ok sends to client a status OK
func (c *Client) Ok(action string) {
	s, err := json.Marshal(struct {
		Action string           `json:"action"`
		Data   SrvStatusMessage `json:"data"`
	}{
		Action: action,
		Data: SrvStatusMessage{
			Status: ErrOK,
			Error:  "OK",
		},
	})
	if !c.CheckError(err, "Can't marhsal message") {
		c.Disconnect()
		return
	}
	c.outgoing <- s
}

// Error sends to client an error status
func (c *Client) Error(action string, text string, status int, closeConn bool) {
	message := SrvStatusMessage{
		Status: status,
		Error:  text,
	}
	data, err := json.Marshal(struct {
		Action string           `json:"action"`
		Data   SrvStatusMessage `json:"data"`
	}{
		Action: action, Data: message,
	})
	if !c.CheckError(err, "Can't marhsal message") {
		c.Disconnect()
		return
	}
	log.Printf("Error: from %v - %v\n", c.ip, text)
	c.writer.Write(data)
	c.writer.Flush()
	if closeConn {
		c.Disconnect()
	}
}

func (c *Client) read() {
	message := SrvWelcomeMessage{
		Action:  "welcome",
		Time:    int(time.Now().Unix()),
		Message: "Happy New Year! Welcome to message server!",
	}
	start, err := json.Marshal(message)
	if !c.CheckError(err, "Can't marhsal message") {
		c.Error("unknown", "Invalid request", ErrInvalidData, true)
		return
	}
	c.outgoing <- start
	log.Printf("Send message to client\n")
	dec := json.NewDecoder(c.reader)
	for {
		var m CltRequest
		err := dec.Decode(&m)
		if !c.CheckError(err, "Invalid message\n") {
			c.Error("unknown", "Invalid request", ErrInvalidData, true)
			return
		}
		log.Printf("Action %v, %v\n", m.Action, string(m.RawData))
		if c.uid == "" && (m.Action != "register" && m.Action != "auth") {
			c.Error(m.Action, "Need auth", ErrNeedAuth, false)
			continue
		}
		switch m.Action {
		case "register":
			var im CltRegister
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Register: Invalid data", ErrInvalidData, true)
				return
			}
			if c.login != "" {
				c.Error(m.Action, "Already register", ErrAlreadyRegister, true)
				return
			}
			c.Register(im.Login, im.Pass, im.Nick)

		case "auth":
			var im CltAuth
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Auth: Invalid data", ErrInvalidData, true)
				return
			}
			if !c.Auth(im.Login, im.Pass) {
				return
			}

		case "setuserinfo":
			var im CltSetUserInfo
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "SetUserInfo: Invalid data", ErrInvalidData, true)
				return
			}
			c.SetUserInfo(im.Avatar, im.Email, im.Phone, im.UserStatus)

		case "userinfo":
			var im CltUserInfo
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Invalid data", ErrInvalidData, true)
				return
			}
			gServer.GetUserInfo(c, im.User)

		case "contactlist":
			var im CltBaseReq
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Channels Invalid data", ErrInvalidData, true)
				return
			}
			c.GetContactList()

		case "addcontact":
			var im CltUidReq
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Invalid data", ErrInvalidData, true)
				return
			}
			c.AddContact(im.User)

		case "delcontact":
			var im CltUidReq
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Invalid data", ErrInvalidData, true)
				return
			}
			c.DelContact(im.User)

		case "message":
			var im CltMessage
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Invalid data", ErrInvalidData, true)
				return
			}
			gServer.SendMessage(c, im.User, im.Body, im.Attach)

		case "import":
			var im CltImport
			err := json.Unmarshal(m.RawData, &im)
			if !c.CheckError(err, "Invalid RawData"+string(m.RawData)) {
				c.Error(m.Action, "Invalid data", ErrInvalidData, true)
				return
			}
			c.ImportContacts(im.Contacts)
		default:
		}
	}
}
