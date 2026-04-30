# A Websocket based chatting app backend written in go

## Features
- Instant messaging with websocket
- User authentication with jwt tokens
- Message editing
- Message limit
- Message deleting

## Todo
- User signup (currently new users should be created manually in the data base)
- Slash commands (there is alreadt >profile to change a users pfp)
- Password hashing
- More secure CORS

## Dependencies
- go-sql-driver/mysql
- golang-jwt/jwt/v5
- golang.org/x/crypto
- mariadb/mysql

## How to use

### Step 1: Configure

rename conf/app_cfg.yml-example to conf/app_cfg.yml
```bash
mv conf/app_cfg.yml-example conf/app_cfg.yml
```
- set your host in "server:host:"
- set your port in "server:port:"
- enable or disable tls in "server:tls:" (requires certificates)
- set the message limit in "server:message_limit:"
- set your mysql database url in "database:mysql:url"
- set your secret in "secret:" (used for jwt tokens)

### Step 2: Initialize Mariadb/Mysql
schema files are located in the schema/ directory
run the message.sql and user.sql files in your mariadb/mysql database

> the denpa.sql file is not necessary it was used in another website cause i didn't want to make a whole other backend for one request

### Step 3: Run
```bash
go run main.go
```

## Documentations

> You have to make a frontend to be able to use this chatapp. here are the endpoints that you need.

### To login:
```
POST /login body: {"username": "username", "password": "password"}
```
This endpoint returns a jwt token if username and password are provided correctly otherwise returns 404, save the token for later use.
### To authenticate:
```
POST /user headers: {Authorization: {token}}
```
This endpoint reads the jwt token in the Authorization header and returns the corresponding user for that token or 403 if the user wasn't found.
### To Receive message history
```
GET /message/{offset} headers: {"Authorization": "token"}
```
This will return a list of {limit} message objects offseted by {limit}*{offset}.
```json
type SendMessage struct {
	"op_code":"op_code",
	"sender":"sender",
	"id":"id",
	"sender_id":"sender_id",
	"content":"content",
	"avatar":"avatar",
	"time":"time"
}
```
- op_code is for websocket and can be ignored.
- sender is the authors username
- id is the message id
- sender_id is the authors id
- content is the message content
- avatar is the authors profile picture url
- time is the local backend time of when the message was authorized 

For example, if your limit in the conf file is 100, /message/0 returns first 100 messages, /message/1 returns 100 messages offset by 100 (2nd 100 messages) etc...
If your token is invalid you will receive a 403 response instead

### Connect to the Websocket
the websocket receives a marshalled json (bytes array) with the following struture.
```json
{
    "token": "jwt token",
    "op_code": "0,1",
    "content": "message to send"
}
```
op code is explained later.

```
WS /ws
```
> After making the connection, your first message sent should be authentication, otherwise the backend will close the connection on the first broadcast from the backend side.
To authenticate you have to send a json object to the websocket with your token and the op_code set to 0 (content can be empty).
> Note that the websocket authentication is seperate from the /user one, /user is used to receive your user data from the backend
### Send message

### WS method:
```json
{
    "token": "jwt token",
    "op_code": "1",
    "content": "message to send"
}
```
Send this object to the websocket with the op_code set to 1.
### http method:
```
POST /message body: {"contet":"message content"} headers: {"Authorization": "token"}
```
> On both of this methods if everything went right, the new message object with an op_code of 1 will be saved to db and broadcasted to all the connected websockets.
### Edit message
```
POST /editmessage body: {"message_id": "message_id", "new_content": "new_content"} headers: {"Authorization": "token"}
```
> If everything goes right, an object will be broadcasted to all the websockets with the following structure
```json
{
    "op_code": 3, // means edit message
    "id": "edited_message_id",
    "new_content": "new_edited_content"
}
```
### Delete message
```
POST /deletemessage body: {"message_id": "message_id"} headers: {"Authorization": "token"}
```
> If everything goes right, an object will be broadcasted to all the websockets with the following structure
```json
{
    "op_code": 2, // means delete message
    "id": "edited_message_id"
}
```
