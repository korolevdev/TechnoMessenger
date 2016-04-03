package server

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"strconv"
	"time"
)

// Error codes
const (
	ErrOK              = 0 // All OK
	ErrAlreadyExist    = 1 // Login or Nickname or Channel already exist
	ErrInvalidPass     = 2 // Invalid login or password
	ErrInvalidData     = 3 // Invalid JSON
	ErrEmptyField      = 4 // Empty Nick, Login, Password or Channel
	ErrAlreadyRegister = 5 // User is already registered
	ErrNeedAuth        = 6 // User has to auth
	ErrNeedRegister    = 7 // User has to register
	ErrUserNotFound    = 8 // User not found by uid
)

///////////////// Server Class ////////////////////////////////////////////////

// Server is an interface of server
type Server interface {
	Start(port int)
	Auth(c *Client, login string, pass string) (string, int, error)
	GetUserData(uid string) (*Client, bool)
	GetUserInfo(c *Client, uid string)
	Register(c *Client, login string, pass string, nick string) (int, error)
	SendMessage(c *Client, uid string, body string, attach AttachData)
	UpdateUserData(c *Client, email string, phone string)
}

// MessageServer is global data storage
type MessageServer struct {
	Logins       map[string]string
	Nicks        map[string]string
	LoginsPasses map[string]string
	Users        map[string]string
	emails       map[string]string // map key - email; val - uid
	phones       map[string]string // map key - phone; val - uid
	Clients      map[string]*Client
}

// NewServer is constructor of Server
func newServer() *MessageServer {
	s := &MessageServer{
		Logins:       make(map[string]string),
		Nicks:        make(map[string]string),
		LoginsPasses: make(map[string]string),
		Users:        make(map[string]string),
		emails:       make(map[string]string),
		phones:       make(map[string]string),
		Clients:      make(map[string]*Client),
	}
	return s
}

var gServer *MessageServer

// CreateInstance create instance of Server
func CreateInstance() Server {
	gServer = newServer()
	return gServer
}

// GetInstance gets instance of Server
func GetInstance() Server {
	return gServer
}

// Start starts server
func (s *MessageServer) Start(port int) {
	psock, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	CheckError(err, "Can't create a server", true)
	log.Printf("Server start on port %v \n", port)
	for {
		conn, err := psock.Accept()
		if CheckError(err, "Can't create connection", false) {
			client := NewClient(conn)
			client.Listen()
		}
	}
}

// Auth checks auth of Client
func (s *MessageServer) Auth(c *Client, login string, pass string) (string, int, error) {
	if login == "" || pass == "" {
		return "", ErrEmptyField, errors.New("Empty field")
	}

	nick, ok := s.Logins[login]
	if !ok {
		return "", ErrNeedRegister, errors.New("Need to register")
	}
	p, ok := s.LoginsPasses[login]
	if !ok || p != pass {
		return "", ErrInvalidPass, errors.New("Invalid login or password!")
	}

	var old *Client
	old, ok = s.Clients[login]
	if ok && old != c {
		old.conn.Close() // Force close connection
		c.status = old.status
		c.avatar = old.avatar
		c.email = old.email
		c.phone = old.phone
		c.contacts = old.contacts
		c.offlineMessages = old.offlineMessages
	}
	s.Clients[login] = c
	c.nick = nick
	c.cid = login

	return GetMD5Hash(login), ErrOK, nil
}

// Register adds new user
func (s *MessageServer) Register(c *Client, login string, pass string, nick string) (int, error) {
	if login == "" || nick == "" || pass == "" {
		return ErrEmptyField, errors.New("Empty field")
	}
	_, ok := s.Nicks[nick]
	if ok {
		return ErrAlreadyExist, errors.New("Nick already was used")
	}
	_, ok = s.Logins[login]
	if ok {
		return ErrAlreadyExist, errors.New("Login already was used")
	}
	s.Nicks[nick] = login
	s.Logins[login] = nick
	s.LoginsPasses[login] = pass
	c.cid = login
	s.Users[c.cid] = login
	return ErrOK, nil
}

// GetUserInfo gets user info to another user
func (s *MessageServer) GetUserInfo(c *Client, uid string) {
	client, ok := s.GetUserData(uid)
	if !ok {
		c.Error("userinfo", "User not found", ErrUserNotFound, false)
		return
	}
	m := SrvUserInfo{
		Nick:       client.nick,
		UserStatus: client.status,
		Email:      client.email,
		Phone:      client.phone,
		Avatar:     client.avatar,
	}
	m.Status = ErrOK
	m.Error = "OK"

	mess, err := json.Marshal(struct {
		Action string      `json:"action"`
		Data   SrvUserInfo `json:"data"`
	}{
		Action: "userinfo",
		Data:   m,
	})

	if !c.CheckError(err, "Can't marhsal message") {
		c.Disconnect()
		return
	}
	c.outgoing <- mess
}

// GetUserData returned user by UserID
func (s *MessageServer) GetUserData(uid string) (*Client, bool) {
	c, ok := s.Clients[uid]
	return c, ok
}

// FindUser finds user by email or phone number
func (s *MessageServer) FindUser(email string, phone string) (*Client, bool) {
	uid, ok := "", false
	if email != "" {
		uid, ok = s.emails[email]
	}
	if !ok && phone != "" {
		uid, ok = s.phones[phone]
	}

	if ok {
		var user *Client
		user, _ = s.GetUserData(uid)
		return user, true
	}

	return nil, false
}

// SendMessage user sends message to channel
func (s *MessageServer) SendMessage(c *Client, uid string, body string, attach AttachData) {
	if body == "" {
		c.Error("message", "Body is empty", ErrEmptyField, false)
		return
	}

	user, ok := s.GetUserData(uid)
	if !ok {
		c.Error("message", "Invalid user", ErrUserNotFound, false)
		return
	}
	c.Ok("message")

	m, err := json.Marshal(struct {
		Action string       `json:"action"`
		Data   EvSrvMessage `json:"data"`
	}{
		Action: "ev_message",
		Data: EvSrvMessage{
			From:   c.cid,
			Nick:   c.nick,
			Body:   body,
			Time:   int(time.Now().Unix()),
			Attach: attach,
		},
	})
	if !c.CheckError(err, "Can't marhsal answer") {
		return
	}
	user.outgoing <- m
	c.outgoing <- m
}

// UpdateUserData - update email and phone
func (s *MessageServer) UpdateUserData(c *Client, email string, phone string) {
	if email != "" {
		if email != c.email && c.email != "" {
			delete(s.emails, c.email)
		}
		c.email = email
		s.emails[email] = c.login
	}

	if phone != "" {
		if phone != c.phone && c.phone != "" {
			delete(s.phones, c.phone)
		}
		c.phone = phone
		s.phones[phone] = c.login
	}
}
