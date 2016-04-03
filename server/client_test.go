package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"testing"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

// NewTestClient creates and inits of Client
func NewTestClient(conn net.Conn) *Client {
	c := NewClient(conn)
	go c.write()
	return c
}

// TestClientDelContact checks Client.DelContact
func TestClientDelContact(t *testing.T) {
	conn := newTestConn()
	c := NewTestClient(conn)
	c.DelContact("111")
	testString := "{\"action\":\"delcontact\",\"data\":{\"status\":0,\"error\":\"OK\"}}"
	err := conn.CheckLastMessage(t, testString)
	if nil != err {
		t.Errorf(err.Error())
	}
}

// TestClientDisconnect checks Client.Disconnect
func TestClientDisconnect(t *testing.T) {
	conn := newTestConn()
	c := NewTestClient(conn)
	c.Disconnect()
	if !conn.Closed {
		t.Errorf("Connection is not closed")
	}
}

// TestClientCheckError checks Client.CheckError
func TestClientCheckError(t *testing.T) {
	conn := newTestConn()
	c := NewTestClient(conn)

	err := errors.New("TestError")
	const str = "Test string"

	if !c.CheckError(nil, str) {
		t.Errorf("CheckError(%v, %s) - return false istead true", nil, str)
	}

	if c.CheckError(err, str) {
		t.Errorf("CheckError(%v, %s) - return true istead false", err, str)
	}
}

// TestClientSetUserInfo checks Client.SetUserInfo
func TestClientSetUserInfo(t *testing.T) {
	gServer = newServer()

	conn := newTestConn()
	c := NewTestClient(conn)

	/*var testData = []struct {
		login, mail, ava, phone, status string
	}{
		{"login", "test@mail.ru", "Base64_Picture", "+7999123123123", "Test State"},
		{"login", "test2@mail.ru", "", "+7999123123777", ""},
	}*/

	c.login = "login"
	mail := "test@mail.ru"
	ava := "Base64_Picture"
	phone := "+7999123123123"
	status := "Test State"

	mail2 := "test2@mail.ru"
	phone2 := "+7999123123777"

	c.SetUserInfo(ava, mail, phone, status)

	if c.email != mail {
		t.Errorf("Email is invalid '%s' instead '%s'", c.email, mail)
	}
	if c.avatar != ava {
		t.Errorf("Avatar is invalid '%s' instead '%s'", c.avatar, ava)
	}
	if c.phone != phone {
		t.Errorf("Phone is invalid '%s' instead '%s'", c.phone, phone)
	}
	if c.status != status {
		t.Errorf("Status is invalid '%s' instead '%s'", c.status, status)
	}

	testString := "{\"action\":\"setuserinfo\",\"data\":{\"status\":0,\"error\":\"OK\"}}"
	err := conn.CheckLastMessage(t, testString)
	if nil != err {
		t.Errorf(err.Error())
	}

	uid, ok := gServer.emails[mail]
	if !ok || uid != c.login {
		t.Errorf("Email was not found (ok = %v) or not valid uid (uid = '%v')", ok, uid)
	}

	uid, ok = gServer.phones[phone]
	if !ok || uid != c.login {
		t.Errorf("Phone was not found (ok = %v) or not valid uid (uid = '%v')", ok, uid)
	}

	c.SetUserInfo("", mail2, phone2, "")
	// TODO: WTF
	c.outgoing <- []byte("")

	err = conn.CheckLastMessage(t, testString)
	if nil != err {
		t.Errorf(err.Error())
	}

	if c.email != mail2 {
		t.Errorf("Email is invalid '%s' instead of '%s'", c.email, mail2)
	}
	if c.avatar != "" {
		t.Errorf("Avatar is invalid '%s' instead of '%s'", c.avatar, "")
	}
	if c.phone != phone2 {
		t.Errorf("Phone is invalid '%s' instead of '%s'", c.phone, phone2)
	}
	if c.status != "" {
		t.Errorf("Status is invalid '%s' instead of '%s'", c.status, "")
	}

	uid, ok = gServer.emails[mail2]
	if !ok || uid != c.login {
		t.Errorf("Email was not found (ok = %v) or not valid uid (uid = '%v')", ok, uid)
	}

	uid, ok = gServer.phones[phone2]
	if !ok || uid != c.login {
		t.Errorf("Phone was not found (ok = %v) or not valid uid (uid = '%v')", ok, uid)
	}

	uid, ok = gServer.emails[mail]
	if ok {
		t.Errorf("Email was found (ok = %v) after deleting\n%v", ok, gServer.emails)
	}

	uid, ok = gServer.phones[phone]
	if ok {
		t.Errorf("Phone was found (ok = %v) after deleting", ok)
	}
}

// TestClientRegister checks Client.Register and Server.Register
func TestClientRegister(t *testing.T) {
	gServer = newServer()

	conn := newTestConn()
	c := NewTestClient(conn)

	testEmpty := "{\"action\":\"register\",\"data\":{\"status\":4,\"error\":\"Empty field\"}}"
	testOk := "{\"action\":\"register\",\"data\":{\"status\":0,\"error\":\"OK\"}}"
	testNickUse := "{\"action\":\"register\",\"data\":{\"status\":1,\"error\":\"Nick already was used\"}}"
	testLoginUse := "{\"action\":\"register\",\"data\":{\"status\":1,\"error\":\"Login already was used\"}}"
	//fmt.Printf("Messages - %v", conn.Messages)

	var testData = []struct {
		login, pass, nick, message string
		conn                       bool
	}{
		{"", "", "", testEmpty, false},
		{"", "1", "1", testEmpty, false},
		{"1", "", "1", testEmpty, false},
		{"1", "1", "", testEmpty, false},
		{"login", "pass", "nick", testOk, true},
		{"login", "pass2", "nick2", testLoginUse, false},
		{"login2", "pass2", "nick", testNickUse, false},
	}

	for _, value := range testData {
		conn.Closed = false
		c.connected = true
		c.Register(value.login, value.pass, value.nick)
		err := conn.CheckLastMessage(t, value.message)
		if nil != err {
			t.Errorf("Register('%s','%s','%s') - '%s'",
				value.login, value.pass, value.nick, err.Error())
		}
		if conn.Closed == value.conn || c.connected != value.conn {
			t.Errorf("Register('%s','%s','%s') Connection has invalid state (%v) instead (%v)!",
				value.login, value.pass, value.nick, conn.Closed, value.conn)
		}
	}
	conn.Closed = false

	var login, nick, pass string
	var ok bool
	login, ok = gServer.Nicks["nick"]
	if !ok || "login" != login {
		t.Errorf("Login was not found (ok = %v) or not valid login ('%v')", ok, login)
	}
	nick, ok = gServer.Logins["login"]
	if !ok || "nick" != nick {
		t.Errorf("Nick was not found (ok = %v) or not valid nick ('%v')", ok, nick)
	}
	pass, ok = gServer.LoginsPasses["login"]
	if !ok || "pass" != pass {
		t.Errorf("Pass was not found (ok = %v) or not valid pass ('%v')", ok, pass)
	}
	login, ok = gServer.Users["login"]
	if !ok || "login" != login {
		t.Errorf("User was not found (ok = %v) or not valid login ('%v')", ok, login)
	}
}

// TestClientAuth checks Client.Auth and Server.Auth
func TestClientAuth(t *testing.T) {
	gServer = newServer()

	testNeedReg := "{\"action\":\"auth\",\"data\":{\"status\":7,\"error\":\"Need to register\"}}"
	testEmpty := "{\"action\":\"auth\",\"data\":{\"status\":4,\"error\":\"Empty field\"}}"
	testLoginFail := "{\"action\":\"auth\",\"data\":{\"status\":2,\"error\":\"Invalid login or password!\"}}"
	testOk := "{\"action\":\"auth\",\"data\":{\"sid\":\"d56b699830e77ba53855679cb1d252da\",\"cid\":\"login\",\"nick\":\"nick\",\"status\":0,\"error\":\"OK\"}}"

	type testfunc func(*Client, *testConn)

	mail := "test@mail.ru"
	ava := "Base64_Picture"
	phone := "+7999123123123"
	status := "Test State"

	setInfo := func(c *Client, conn *testConn) {
		c.SetUserInfo(ava, mail, phone, status)
		c.outgoing <- []byte("")
		if c.email != mail {
			t.Errorf("Email is invalid '%s' instead '%s'", c.email, mail)
		}
		if c.avatar != ava {
			t.Errorf("Avatar is invalid '%s' instead '%s'", c.avatar, ava)
		}
		if c.phone != phone {
			t.Errorf("Phone is invalid '%s' instead '%s'", c.phone, phone)
		}
		if c.status != status {
			t.Errorf("Status is invalid '%s' instead '%s'", c.status, status)
		}
		conn.ClearMessages()
	}
	checkReconect := func(c *Client, _ *testConn) {
		if c.email != mail {
			t.Errorf("Email is invalid '%s' instead '%s'", c.email, mail)
		}
		if c.avatar != ava {
			t.Errorf("Avatar is invalid '%s' instead '%s'", c.avatar, ava)
		}
		if c.phone != phone {
			t.Errorf("Phone is invalid '%s' instead '%s'", c.phone, phone)
		}
		if c.status != status {
			t.Errorf("Status is invalid '%s' instead '%s'", c.status, status)
		}
	}

	var testData = []struct {
		login, pass, nick, mess    string
		connect, register, newconn bool
		prefunc, postfunc          testfunc
	}{
		{"", "", "", testEmpty, false, false, true, nil, nil},
		{"login", "", "", testEmpty, false, false, true, nil, nil},
		{"", "pass", "", testEmpty, false, false, true, nil, nil},
		{"login", "pass", "nick", testNeedReg, false, false, true, nil, nil},
		{"login", "pass", "nick", testOk, true, true, true, nil, nil},
		{"login", "saap", "nick", testLoginFail, false, false, true, nil, nil},
		{"login", "pass", "nick", testOk, true, false, true, nil, setInfo},
		{"login", "pass", "nick", testOk, true, false, true, nil, checkReconect},
	}

	var conn *testConn
	var c *Client
	for _, val := range testData {
		if val.newconn {
			conn = newTestConn()
			c = NewTestClient(conn)
		}

		if val.register {
			//t.Logf("TestClientAuth - gServer.Register(%v, %v, %v)", val.login, val.pass, val.nick)
			gServer.Register(c, val.login, val.pass, val.nick)
		}

		if val.prefunc != nil {
			val.prefunc(c, conn)
		}

		c.Auth(val.login, val.pass)
		err := conn.CheckLastMessage(t, val.mess)
		if nil != err {
			t.Errorf("Test data - (%v): %s", val, err.Error())
		}
		if conn.Closed == val.connect || c.connected != val.connect {
			t.Errorf("Test data - (%v): connection invalid (conn - %v, client - %v)", val, conn.Closed, c.connected)
		}
		if val.postfunc != nil {
			val.postfunc(c, conn)
		}
	}
}

func TestClientAddContact(t *testing.T) {
	gServer = newServer()

	conn := newTestConn()
	c1 := NewTestClient(newTestConn())
	c := NewTestClient(conn)

	testOk := "{\"action\":\"auth\",\"data\":{\"sid\":\"d56b699830e77ba53855679cb1d252da\",\"cid\":\"login\",\"nick\":\"nick\",\"status\":0,\"error\":\"OK\"}}"

	gServer.Register(c1, "user", "pass", "user1")
	gServer.Register(c, "login", "pass", "nick")
	c1.Auth("user", "pass")
	c.Auth("login", "pass")
	c.outgoing <- []byte("")
	err := conn.CheckLastMessage(t, testOk)
	if nil != err {
		t.Errorf(err.Error())
	}

	ansOk := "{\"action\":\"addcontact\",\"data\":{\"status\":0,\"error\":\"OK\"}}"
	ansAlready := "{\"action\":\"addcontact\",\"data\":{\"status\":1,\"error\":\"User already in list\"}}"
	ansNotFound := "{\"action\":\"addcontact\",\"data\":{\"status\":8,\"error\":\"User not found\"}}"

	c.AddContact(c1.login)
	c.outgoing <- []byte("")
	err = conn.CheckLastMessage(t, ansOk)
	if nil != err {
		t.Errorf(err.Error())
	}

	c.AddContact(c.login)
	c.outgoing <- []byte("")
	err = conn.CheckLastMessage(t, ansAlready)
	if nil != err {
		t.Errorf(err.Error())
	}

	c.AddContact(c1.login)
	err = conn.CheckLastMessage(t, ansAlready)
	if nil != err {
		t.Errorf(err.Error())
	}

	c.AddContact("unknown")
	err = conn.CheckLastMessage(t, ansNotFound)
	if nil != err {
		t.Errorf(err.Error())
	}
}

// TestClientGetContactList checks Client.GetContactList
func TestClientGetContactList(t *testing.T) {
	gServer = newServer()

	testUsers := []struct {
		conn              *testConn
		client            *Client
		login, pass, nick string
		email, phone, ava string
	}{}

	for i := 0; i < 1; i++ {
		conn := newTestConn()
		tmp := struct {
			conn              *testConn
			client            *Client
			login, pass, nick string
			email, phone, ava string
		}{
			conn,
			NewTestClient(conn),
			fmt.Sprintf("user%v", i),
			fmt.Sprintf("pass%v", i),
			fmt.Sprintf("nick%v", i),

			fmt.Sprintf("mail%v@mail.ru", i),
			fmt.Sprintf("+6722%v", i),
			fmt.Sprintf("Ava_Picture%v", i),
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

		tmp.client.SetUserInfo(tmp.ava, tmp.email, tmp.phone, "")
	}

	conn := newTestConn()
	c := NewTestClient(conn)
	gServer.Register(c, "login", "pass", "nick")
	c.Auth("login", "pass")

	ansUserTmpl := "{\"uid\":\"%s\",\"nick\":\"%s\",\"email\":\"%s\",\"phone\":\"%s\",\"picture\":\"%s\"},"

	messUsers := ""
	for _, val := range testUsers {
		c.AddContact(val.client.uid)
		messUsers += fmt.Sprintf(ansUserTmpl, val.login, val.nick, val.email, val.phone, val.ava)
	}

	andOkTml := "{\"action\":\"contactlist\",\"data\":{\"list\":[%s],\"status\":0,\"error\":\"OK\"}}"
	mess := fmt.Sprintf(andOkTml, messUsers[:len(messUsers)-1])
	c.GetContactList()
	c.outgoing <- []byte("")
	err := conn.CheckLastMessage(t, mess)
	if nil != err {
		t.Errorf(err.Error())
	}
}

// TestClientImportContacst checks Client.ImportContacts
func TestClientImportContacst(t *testing.T) {
	gServer = newServer()

	testUsers := []struct {
		conn              *testConn
		client            *Client
		login, pass, nick string
		email, phone, ava string
	}{}

	for i := 0; i < 10; i++ {
		conn := newTestConn()
		tmp := struct {
			conn              *testConn
			client            *Client
			login, pass, nick string
			email, phone, ava string
		}{
			conn,
			NewTestClient(conn),
			fmt.Sprintf("user%v", i),
			fmt.Sprintf("pass%v", i),
			fmt.Sprintf("nick%v", i),

			fmt.Sprintf("mail%v@mail.ru", i),
			fmt.Sprintf("+6722%v", i),
			fmt.Sprintf("Ava_Picture%v", i),
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
		tmp.client.SetUserInfo(tmp.ava, tmp.email, tmp.phone, "")
	}

	conn := newTestConn()
	c := NewTestClient(conn)
	gServer.Register(c, "login", "pass", "nick")
	c.Auth("login", "pass")

	ansUserTmpl := "{\"uid\":\"%s\",\"nick\":\"%s\",\"email\":\"%s\",\"phone\":\"%s\",\"picture\":\"%s\",\"myid\":\"%s\"},"

	var testContacts = make([]Contact, 0)
	testContacts = append(testContacts, Contact{"MyName1", "+67228", "akjdhf", "1"})
	testContacts = append(testContacts, Contact{"MyName2", "777", "mail2@mail.ru", "2"})
	testContacts = append(testContacts, Contact{"MyName3", "a", "adfj", "3"})
	testContacts = append(testContacts, Contact{"MyName4", "b", "sda", "4"})
	testContacts = append(testContacts, Contact{"MyName5", "98", "ffkhh", "5"})
	testContacts = append(testContacts, Contact{"MyName6", "sdaf", "ff", "6"})
	testContacts = append(testContacts, Contact{"MyName7", "", "", "7"})
	testContacts = append(testContacts, Contact{"MyName8", "df", "df", "8"})

	c.ImportContacts(testContacts)

	messUsers := ""
	messUsers += fmt.Sprintf(ansUserTmpl, testUsers[8].login,
		testUsers[8].nick, testUsers[8].email,
		testUsers[8].phone, testUsers[8].ava, "1")
	messUsers += fmt.Sprintf(ansUserTmpl, testUsers[2].login,
		testUsers[2].nick, testUsers[2].email,
		testUsers[2].phone, testUsers[2].ava, "2")

	andOkTml := "{\"action\":\"import\",\"data\":{\"list\":[%s],\"status\":0,\"error\":\"OK\"}}"
	mess := fmt.Sprintf(andOkTml, messUsers[:len(messUsers)-1])
	c.outgoing <- []byte("")
	err := conn.CheckLastMessage(t, mess)
	if nil != err {
		t.Errorf(err.Error())
	}
}
