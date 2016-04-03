package server

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

/// Mock Objects (BEGIN)///////////////////////////////////////////////////////

// testAddr - Mock net.Addr
type testAddr struct {
	id string
}

// Network - name of the network
func (t *testAddr) Network() string {
	return "Network " + t.id
}

// String - string form of address
func (t *testAddr) String() string {
	return "String " + t.id
}

// testConn - Mock net.Conn
type testConn struct {
	Messages   []string
	Closed     bool
	localAddr  testAddr
	remoteAddr testAddr
}

// newTestConn - constructor of testConn
func newTestConn() *testConn {
	conn := &testConn{
		Messages:   make([]string, 0),
		Closed:     false,
		localAddr:  testAddr{id: "Local"},
		remoteAddr: testAddr{id: "Remote"},
	}
	return conn
}

// LocalAddr returns the local network address.
func (c *testConn) LocalAddr() net.Addr {
	return &c.localAddr
}

// RemoteAddr returns the remote network address.
func (c *testConn) RemoteAddr() net.Addr {
	return &c.remoteAddr
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *testConn) Close() error {
	c.Closed = true
	return nil
}

// Read reads data from the connection.
// Read can be made to time out and return a Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *testConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

// Write writes data to the connection.
// Write can be made to time out and return a Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *testConn) Write(b []byte) (n int, err error) {
	c.Messages = append(c.Messages, string(b[:len(b)]))
	return len(b), nil
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future I/O, not just
// the immediately following call to Read or Write.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *testConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls.
// A zero value for t means Read will not time out.
func (c *testConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *testConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// ClearMessages clears of message history
func (c *testConn) ClearMessages() {
	c.Messages = make([]string, 0)
}

// CheckLastMessage checks last message
func (c *testConn) CheckLastMessage(t *testing.T, mess string) error {

	if mess == "" && len(c.Messages) == 0 {
		return nil
	}

	if len(c.Messages) == 0 {
		return errors.New("Empty messages")
	}
	str := c.Messages[len(c.Messages)-1]
	if str != mess {
		return fmt.Errorf("Not correct answer '%s' waits - '%s'", str, mess)
	}
	c.Messages = c.Messages[:len(c.Messages)-1]
	return nil
}

/// Mock Objects (END)/////////////////////////////////////////////////////////

////////// Server tests //////////////////////////////////////////////////////

// TestServerGetUserInfo checks Server.GetUserInfo
func TestServerGetUserInfo(t *testing.T) {
	gServer = newServer()

	conn := newTestConn()
	c1 := NewTestClient(newTestConn())
	c := NewTestClient(conn)

	gServer.Register(c1, "user", "pass", "user1")
	gServer.Register(c, "login", "pass", "nick")
	mail := "test@mail.ru"
	ava := "Base64_Picture"
	phone := "+7999123123123"
	status := "Test State"

	ansOk := "{\"action\":\"userinfo\",\"data\":{\"nick\":\"user1\",\"user_status\":\"Test State\",\"email\":\"test@mail.ru\",\"phone\":\"+7999123123123\",\"picture\":\"Base64_Picture\",\"status\":0,\"error\":\"OK\"}}"
	ansNotFound := "{\"action\":\"userinfo\",\"data\":{\"status\":8,\"error\":\"User not found\"}}"

	c1.Auth("user", "pass")
	c1.SetUserInfo(ava, mail, phone, status)

	gServer.GetUserInfo(c, c1.uid)
	c.outgoing <- []byte("")

	err := conn.CheckLastMessage(t, ansOk)
	if nil != err {
		t.Errorf(err.Error())
	}

	gServer.GetUserInfo(c, "unknown")
	err = conn.CheckLastMessage(t, ansNotFound)
	if nil != err {
		t.Errorf(err.Error())
	}
}

// TestServerFindUser checks Server.FindUser
func TestServerFindUser(t *testing.T) {
	gServer = newServer()

	c1 := NewTestClient(newTestConn())
	c2 := NewTestClient(newTestConn())

	gServer.Register(c1, "user1", "pass", "user1")
	gServer.Register(c2, "user2", "pass", "user2")
	c1.Auth("user1", "pass")
	c2.Auth("user2", "pass")
	mail1 := "test1@mail.ru"
	phone1 := "+7999123123123"

	//TODO: Check same emails
	mail2 := "test2@mail.ru"
	phone2 := "+7999123123124"

	c1.SetUserInfo("", mail1, phone1, "")
	c2.SetUserInfo("", mail2, phone2, "")

	c, ok := gServer.FindUser("", "")
	if ok {
		t.Errorf("Found empty user")
	}
	c, ok = gServer.FindUser(mail1, phone1)
	if !ok || c == nil || c != c1 {
		t.Errorf("Found (%v) invalid user - (%v) waits (%v) ", ok, c, c1)
	}
	c, ok = gServer.FindUser(mail1, "")
	if !ok || c == nil || c != c1 {
		t.Errorf("Found (%v) invalid user - (%v) waits (%v) ", ok, c, c1)
	}
	c, ok = gServer.FindUser("", phone1)
	if !ok || c == nil || c != c1 {
		t.Errorf("Found (%v) invalid user - (%v) waits (%v) ", ok, c, c1)
	}
	c, ok = gServer.FindUser(mail1, phone2)
	if !ok || c == nil || c != c1 {
		t.Errorf("Found (%v) invalid user - (%v) waits (%v) ", ok, c, c1)
	}
}

// TestServerGetUserData checks Server.GetUserData
func TestServerGetUserData(t *testing.T) {
	gServer = newServer()

	conn := newTestConn()
	c := NewTestClient(conn)

	gServer.Register(c, "user", "pass", "user1")
	c.Auth("user", "pass")

	// Not found
	client, ok := gServer.GetUserData("test")
	if ok {
		t.Errorf("Found invalid user")
	}

	client, ok = gServer.GetUserData(c.login)
	if !ok || client == nil {
		t.Errorf("Not found valid user")
	}

	if ok && client != c {
		t.Errorf("Not found valid user %v waits %v", client, c)
	}
}

// TestServerSendMessage check Server.SendMessage
func TestServerSendMessage(t *testing.T) {
	gServer = newServer()

	testUsers := []struct {
		conn              *testConn
		client            *Client
		login, pass, nick string
	}{}

	for i := 0; i < 3; i++ {
		conn := newTestConn()
		tmp := struct {
			conn              *testConn
			client            *Client
			login, pass, nick string
		}{
			conn,
			NewTestClient(conn),
			fmt.Sprintf("user%v", i),
			fmt.Sprintf("pass%v", i),
			fmt.Sprintf("nick%v", i),
		}
		testUsers = append(testUsers, tmp)

		gServer.Register(tmp.client, tmp.login, tmp.pass, tmp.nick)
		tmp.client.Auth(tmp.login, tmp.pass)

		testOk := fmt.Sprintf(
			"{\"action\":\"auth\",\"data\":{\"sid\":\"%s\",\"cid\":\"%s\",\"nick\":\"%s\",\"status\":0,\"error\":\"OK\"}}",
			GetMD5Hash(tmp.login), tmp.login, tmp.nick)

		err := tmp.conn.CheckLastMessage(t, testOk)
		if nil != err {
			t.Errorf("Auth('%s','%s') - '%s'",
				tmp.login, tmp.pass, err.Error())
		}
	}

	c1 := testUsers[0]
	c2 := testUsers[1]
	c3 := testUsers[1]

	testAttaches := []AttachData{
		{"", ""},
		{"", "Test"},
		{"Test", ""},
		{"txt", "Sample Text"},
		{"img", "lkaj;fkladfkljsdlfjs;dlfkj;salfdj;asldfj"},
	}

	testMess := "Test Message Body"
	ansOk := "{\"action\":\"message\",\"data\":{\"status\":0,\"error\":\"OK\"}}"
	ansEmpy := "{\"action\":\"message\",\"data\":{\"status\":4,\"error\":\"Body is empty\"}}"
	ansInvUser := "{\"action\":\"message\",\"data\":{\"status\":8,\"error\":\"Invalid user\"}}"
	ansMessTmpl := "{\"action\":\"ev_message\",\"data\":{\"from\":\"%s\",\"nick\":\"%s\",\"body\":\"%s\",\"time\":%v,\"attach\":{\"mime\":\"%s\",\"data\":\"%s\"}}}"

	// Check empty body
	gServer.SendMessage(c1.client, c2.client.uid, "", testAttaches[0])
	err := c1.conn.CheckLastMessage(t, ansEmpy)
	if nil != err {
		t.Errorf(err.Error())
	}
	if c1.conn.Closed {
		t.Errorf("Connection was closed")
	}

	// Check invalid user
	gServer.SendMessage(c1.client, "invalid", testMess, testAttaches[0])
	err = c1.conn.CheckLastMessage(t, ansInvUser)
	if nil != err {
		t.Errorf(err.Error())
	}
	if c1.conn.Closed {
		t.Errorf("Connection was closed")
	}

	// Check normal message to online
	gServer.SendMessage(c1.client, c2.client.uid, testMess, testAttaches[0])

	mess := fmt.Sprintf(ansMessTmpl, c1.login, c1.nick, testMess, int(time.Now().Unix()),
		testAttaches[0].Mime, testAttaches[0].Data)
	err = c1.conn.CheckLastMessage(t, mess)
	if nil != err {
		t.Errorf(err.Error())
	}
	err = c2.conn.CheckLastMessage(t, mess)
	if nil != err {
		t.Errorf(err.Error())
	}
	err = c1.conn.CheckLastMessage(t, ansOk)
	if nil != err {
		t.Errorf(err.Error())
	}

	gServer.SendMessage(c1.client, c2.client.uid, testMess, testAttaches[1])
	mess = fmt.Sprintf(ansMessTmpl, c1.login, c1.nick, testMess, int(time.Now().Unix()),
		testAttaches[1].Mime, testAttaches[1].Data)
	err = c1.conn.CheckLastMessage(t, mess)
	if nil != err {
		t.Errorf(err.Error())
	}
	err = c2.conn.CheckLastMessage(t, mess)
	if nil != err {
		t.Errorf(err.Error())
	}
	err = c1.conn.CheckLastMessage(t, ansOk)
	if nil != err {
		t.Errorf(err.Error())
	}

	gServer.SendMessage(c1.client, c2.client.uid, testMess, testAttaches[2])
	mess = fmt.Sprintf(ansMessTmpl, c1.login, c1.nick, testMess, int(time.Now().Unix()),
		testAttaches[2].Mime, testAttaches[2].Data)
	err = c1.conn.CheckLastMessage(t, mess)
	if nil != err {
		t.Errorf(err.Error())
	}
	err = c2.conn.CheckLastMessage(t, mess)
	if nil != err {
		t.Errorf(err.Error())
	}
	err = c1.conn.CheckLastMessage(t, ansOk)
	if nil != err {
		t.Errorf(err.Error())
	}

	// Check normal message to offline
	c3.client.Disconnect()
	mess = fmt.Sprintf(ansMessTmpl, c1.login, c1.nick, testMess, int(time.Now().Unix()),
		testAttaches[3].Mime, testAttaches[3].Data)

	gServer.SendMessage(c1.client, c2.client.uid, testMess, testAttaches[3])
	err = c1.conn.CheckLastMessage(t, mess)
	if nil != err {
		t.Errorf(err.Error())
	}
	if string(c3.client.offlineMessages[0]) != mess {
		t.Errorf("Invalid ofline message (%v) instead (%v)", c3.client.offlineMessages, mess)
	}

	conn := newTestConn()
	c := NewTestClient(conn)
	c.Auth(c3.login, c3.pass)
	c.outgoing <- []byte("")
	err = conn.CheckLastMessage(t, mess)
	if nil != err {
		t.Errorf(err.Error())
	}

	if 0 != len(c.offlineMessages) {
		t.Errorf("Offline messages didn't clear")
	}
}
