# Протокол работы с сервером  
## Запуск сервера
go run main.go (localhost:7788)
Подключение по ip локальной wi-fi сети
## Запросы от клиента на сервер  
1. Регистрация
```json
{
	"action":"register",
	"data":{
		"login":"MY_LOGIN",
		"pass":"MD5_FROM_PASS",
		"nick":"NICKNAME"
	}
}
```
2. Авторизация
```json
{
	"action":"auth",
	"data":{
		"login":"MY_LOGIN",
		"pass":"MD5_FROM_PASS"
	}
}
```
3. Запросить информацию о пользователе
```json
{
	"action":"userinfo",
	"data": {
		"user":"USER_ID",
		"cid":"MY_USER_ID",
		"sid":"MY_SESSION_ID"
	}
}
```
4. Запрос контакт листа
```json
{
	"action":"contactlist", 
	"data": {
        "cid":"MY_USER_ID",
        "sid":"MY_SESSION_ID"
    }
}
```
5. Добавление в контакт лист
```json
{
    "action":"addcontact", 
    "data": {
        "uid":"USER_ID",
        "cid":"MY_USER_ID",
        "sid":"MY_SESSION_ID"
    }
}
```
6. Удаление из контакт листа 
```json
{
    "action":"delcontact", 
    "data": {
        "uid":"USER_ID",
        "cid":"MY_USER_ID",
        "sid":"MY_SESSION_ID"
    }
}
```
7. Отправка сообщения 
```json
{
    "action":"message",
    "data": {
        "cid":"MY_USER_ID",
        "sid":"MY_SESSION_ID",
        "uid":"USER_ID",
        "body":"MESSAGE",
        "attach": {
            "mime":"MIME_TYPE_OF_ATTACH",
            "data":"BASE64_OF_ATTACH"
        }
    }
}
```
8. Импорт контактов
```json 
{
    "action":"import",
    "data":{
        "contacts":[
            {
                "myid":"MY_ID",
                "name":"NAME",
                "phone":"PNONE",
                "email":"EMAIL"
            },
            {
                "myid":"MY_ID",
                "name":"NAME",
                "phone":"PNONE",
                "email":"EMAIL" 
            }
        ]
    }
}
```
9. Изменить свою информацию 
```json
{
    "action":"setuserinfo",
    "data": {
        "user_status":"STATUS_STRING",
        "cid":"MY_USER_ID",
        "sid":"MY_SESSION_ID"
        "email":"EMAIL",
        "phone":"PHONE",
        "picture":"BASE64_SMALL_PIC"
    }
 }

## Ответы сервера на клиент
1. Welcome сообщение приходит при конекте к серверу
```json
{
	"action":"welcome",
	"message": "WELCOME_TEXT",
	"time":UNIXTIMESTAMP
}
```
2. Ответ на авторизацию
```json
{
	"action":"auth",
	"data":{
		"status":[0-9]+,
		"error":"TEXT_OF_ERROR",
		"sid":"SESSION_ID",
		"uid":"USER_ID"
	}
}
```
3. Ответ на регистрацию
```json
{
	"action":"register",
	"data":{
		"status":[0-9]+,
		"error":"TEXT_OF_ERROR"
	}
}
```
4. Ответ на запрос информации о пользователе
```json
{
	"action":"userinfo",
	"data":{
		"status":[0-9]+,
		"error":"TEXT_OF_ERROR",
		"nick":"NICKNAME",
        "email":"EMAIL",
        "phone":"PHONE",
        "picture":"BASE64_SMALL_PIC"
		"user_status":"STATUS_STRING"
	}
}
```
5. Ответ на запрос контакт листа
```json
{
    "action":"contactlist", 
    "data":{
        "status":"[0-9]+",
        "error":"TEXT_OF_ERROR",
        "list":[
            {
                "myid":"YOUR_ID",
                "uid":"UID",
                "nick":"NICK NAME",
                "email":"EMAIL",
                "phone":"PHONE",
                "picture":"BASE64_SMALL_PIC"
            },
            {
                "myid":"YOUR_ID",
                "uid":"UID",
                "nick":"NICK NAME",
                "email":"EMAIL",
                "phone":"PHONE",
                "picture":"BASE64_SMALL_PIC"
            },
        ]
    }
}
```
6. Добавление контакта 
```json
{
    "action":"addcontact", 
    "data": {
        "status":"[0-9]+",
        "error":"TEXT_OF_ERROR",
        "user":{
                "uid":"UID",
                "nick":"NICK NAME",
                "email":"EMAIL",
                "phone":"PHONE"
        }
    }
}
```
7. Удаление контакта 
```json
{
    "action":"delcontact", 
    "data": {
        "status":"[0-9]+",
        "error":"TEXT_OF_ERROR",
        "uid":"UID"
    }
}
```
8. Импорт контактов
```json
{
    "action":"import", 
    "data":{
        "status":"[0-9]+",
        "error":"TEXT_OF_ERROR",
        "list":[
            {
                "myid":"YOUR_ID",
                "uid":"UID",
                "nick":"NICK NAME",
                "email":"EMAIL",
                "phone":"PHONE"
            },
            {
                "myid":"YOUR_ID",
                "uid":"UID",
                "nick":"NICK NAME",
                "email":"EMAIL",
                "phone":"PHONE"
            },
        ]
    }
}
```
9. Изменить свою информацию 
```json
{
    "action":"setuserinfo",
    "data":{
        "status":[0-9]+,
        "error":"TEXT_OF_ERROR"
    }
 } 
```
10. Отправка сообщения 
```json
{
    "action":"message", 
    "data":{
        "status":"[0-9]+",
        "error":"TEXT_OF_ERROR"
    }
}
```

## События присылаемые с сервера на клиент
1. Новое сообщение 
```json
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
```

## Коды ошибок 
```golang
// Error codes
const (
	ErrOK              = 0 // All OK
	ErrAlreadyExist    = 1 // Login or Nickname already exist
	ErrInvalidPass     = 2 // Invalid login or password
	ErrInvalidData     = 3 // Invalid JSON
	ErrEmptyField      = 4 // Empty Nick, Login, Password or Channel
	ErrAlreadyRegister = 5 // User is already registered
	ErrNeedAuth        = 6 // User has to auth
	ErrNeedRegister    = 7 // User has to register
	ErrUserNotFound    = 8 // User not found by uid
)
```
