package server

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
)

// CheckError checks errors and print log
func CheckError(err error, message string, fatal bool) bool {
	if err != nil {
		if fatal {
			log.Fatalln(message + ": " + err.Error())
		} else {
			log.Println(message + ": " + err.Error())
		}
	}
	return err == nil
}

// GetMD5Hash calculates MD5
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

//// Client messages

type CltRequest struct {
	Action  string          `json:"action"`
	RawData json.RawMessage `json:"data,omitempty"`
}

type CltAuth struct {
	Login string `json:"login"`
	Pass  string `json:"pass"`
}

type CltRegister struct {
	Nick string `json:"nick"`
	CltAuth
}

type CltBaseReq struct {
	Cid string `json:"cid"`
	Sid string `json:"sid"`
}

type CltUserInfo struct {
	User string `json:"user"`
	CltBaseReq
}

type CltUidReq struct {
	User string `json:"uid"`
	CltBaseReq
}

type CltSetUserInfo struct {
	UserStatus string `json:"user_status"`
	Avatar     string `json:"picture,omitempty"`
	Email      string `json:"email,omitempty"`
	Phone      string `json:"phone,omitempty"`
	CltBaseReq
}

type CltChannel struct {
	Channel string `json:"channel"`
	CltBaseReq
}

type CltCreateChannel struct {
	Name  string `json:"name"`
	Descr string `json:"descr"`
}

type AttachData struct {
	Mime string `json:"mime"`
	Data string `json:"data"`
}

type CltMessage struct {
	Body   string     `json:"body"`
	Attach AttachData `json:"attach,omitempty"`
	CltUidReq
}

type Contact struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	MyID  string `json:"myid,omitempty"`
}

type CltImport struct {
	Contacts []Contact `json:"contacts"`
	CltBaseReq
}

//// Server messages

type SrvMessage struct {
	Action  string          `json:"action"`
	Time    int             `json:"time"`
	RawData json.RawMessage `json:"data,omitempty"`
}

type SrvWelcomeMessage struct {
	Message string `json:"message"`
	Action  string `json:"action"`
	Time    int    `json:"time"`
}

type SrvStatusMessage struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

type SrvAddChannelMessage struct {
	ChannelID string `json:"chid"`
	SrvStatusMessage
}

type SrvStatusAuthMessage struct {
	Sid  string `json:"sid"`
	Cid  string `json:"cid"`
	Nick string `json:"nick"`
	SrvStatusMessage
}

type ChannelData struct {
	Id     string `json:"chid"`
	Name   string `json:"name"`
	Descr  string `json:"descr"`
	Online int    `json:"online"`
}

type SrvUserInfo struct {
	Nick       string `json:"nick"`
	UserStatus string `json:"user_status"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Avatar     string `json:"picture"`
	SrvStatusMessage
}

type UserData struct {
	Uid    string `json:"uid"`
	Nick   string `json:"nick"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Avatar string `json:"picture"`
	MyID   string `json:"myid,omitempty"`
}

type SrvListOfUsers struct {
	Users []UserData `json:"list"`
	SrvStatusMessage
}

type MessageData struct {
	Mid  string `json:"mid"`
	From string `json:"from"`
	Nick string `json:"nick"`
	Body string `json:"body"`
	Time int    `json:"time"`
}

/*
{
	"action":"ev_message",
	"data":{
		"from":"USER_ID",
		"nick":"NICKNAME",
		"body":"TEXT_OF_MESSAGE",
		"time":"TIMESPAMT",
		"attach": {
			"mime":"MIME_TYPE_OF_ATTACH",
			"data":"BASE64_OF_ATTACH"
		}
	}
}
*/
type EvSrvMessage struct {
	From   string     `json:"from"`
	Nick   string     `json:"nick"`
	Body   string     `json:"body"`
	Time   int        `json:"time"`
	Attach AttachData `json:"attach,omitempty"`
}
