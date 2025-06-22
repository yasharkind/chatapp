package main

import (
	"chatapp/base/appconfig"
	"chatapp/dto"
	"chatapp/objects"
	service "chatapp/services"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

const messageFile = "messages.txt"
const messageLimit = 500
var users = map[string]string{
	"hoogMEH": "meowmeow",
	"MhZmn": "banana",
	"seeb": "baka",
	"torob": "baka",
	"theadonn": "denpa",
	"*sneezes*": "oneeeeee",
	"rad": "koko",
	"yishay": "fishe",
	"beh": "koishifumo",
	"looghMEH": "ali.1385",
	"Legendarybtw": "Pass138580",
	"Marisha": "younes12",
	"H3x7": "h3xanol",
	"Adib": "?!?!?!?!",
	"Yoi": "ioyyoi",
	"qonoeba": "konobeba",
}

func validateUser(name string, password string, content string) bool {
	//validate user
	if (users[name] != password || name == ""){
		
		unknownUser := fmt.Sprintf("unknwos user: %s, password: %s\nmessage: %s", name, password,  content)
		fmt.Println(unknownUser)
		f, err := os.OpenFile("unknownuser.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			f.WriteString(unknownUser + "\n")
			f.Close()
		} else {
		fmt.Println("failed to write to file:", err)
		}
		return false
	}
	return true
}

type Server struct {
	conns map[*websocket.Conn]bool

	messageList []*objects.Message

	mu sync.Mutex

	messageService service.MessageService
}

func NewServer(cfg *appconfig.Config) *Server {
	s := &Server{
		conns: make(map[*websocket.Conn]bool),
		messageService: service.NewMessageService(cfg),
	}

	s.messageList = s.messageService.FindFromEnd(10)
//
//	data, err := os.ReadFile(messageFile)
//	if err == nil {
//		s.messageList = append(s.messageList, splitLines(string(data))...)
//	} else {
//		fmt.Println("No message history loaded: ", err)
//	}


	return s
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func (s *Server) handleWS(ws *websocket.Conn) {
	var address = ws.RemoteAddr().String()
	if (!(strings.Contains(address, "touhou.ir") || strings.Contains(address, "localhost"))){
		ws.Close()
		return
	}
	fmt.Println("new incoming connection from client: ", ws.RemoteAddr())
	
	s.mu.Lock()
	s.conns[ws] = true
	s.mu.Unlock()

	s.mu.Lock()
	var slice int
	var mslen = len(s.messageList)
	var full_msg string = "["

	if  mslen < messageLimit { slice = mslen } else { slice = messageLimit }
	for _, message := range s.messageList[mslen - slice:] {
		marshaled, err := json.Marshal(message) 
		if err != nil {
			fmt.Println("error marshaling: ", err)
			return
		}
		full_msg += string(marshaled) + "," 
	}
	full_msg = full_msg[:len(full_msg)-1] + "]"
	if _, err := ws.Write([]byte(full_msg)); err != nil {
		fmt.Println("error sending past messages: ", err)
	}
	s.mu.Unlock()

	s.readLoop(ws)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {

	file, handler, err := r.FormFile("file")
	name := r.FormValue("sender")
	password := r.FormValue("password")

	if (!validateUser(name, password, "file object")){
		return
	}

	if err != nil {
		fmt.Println("formfile error: ", err)
		return 
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	err = os.WriteFile("uploads/" + handler.Filename, content, 0644)
	if err != nil {
		fmt.Println("write error: ", err)
	}

	msg := objects.Message {
		Id: 0,
		Sender: name,
		Content: "http://chat.touhou.ir:3000/files/" + handler.Filename,
		Time:  time.Now(),
	}

	
	s.mu.Lock()
	s.messageList = append(s.messageList, &msg)
	newmsg := s.messageService.Save(msg)
	fmt.Println("uplaoded file: ", handler.Filename)
	marshaled, err := json.Marshal(newmsg)
	if err != nil {
		fmt.Println("marshal err: ", err)	
	}
//	f, err := os.OpenFile(messageFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//	if err == nil {
//		f.WriteString(string(marshaled) + "\n")
//		f.Close()
//	} else {
//		fmt.Println("failed to write to file:", err)
//	}
	s.mu.Unlock()
	s.broadcast([]byte(marshaled))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	decoder := json.NewDecoder(ws)
	for {
		var msg dto.Message
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("read error:", err)
			break
		}

		if (!validateUser(msg.Sender, msg.Password, msg.Content)) {
			return
		}
		
		jsonMsg := fmt.Sprintf("%s-%s: %s", time.Now(), msg.Sender, msg.Content)
		fmt.Println(jsonMsg)
		message := &objects.Message{
			Id: 0,
			Sender: msg.Sender,
			Time: time.Now(),
			Content: msg.Content,
		}

		
	
		s.mu.Lock()
		savedmsg := s.messageService.Save(*message)
		s.messageList = append(s.messageList, savedmsg)
		marshaled, err := json.Marshal(savedmsg)
		if err != nil {
			fmt.Println("Error marshaling: ", err)
		}
	//	f, err := os.OpenFile(messageFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	//	if err == nil {
	//		f.WriteString(string(marshaled) + "\n")
	//		f.Close()
	//	} else {
	//		fmt.Println("failed to write to file:", err)
	//	}
		s.mu.Unlock()

		s.broadcast([]byte(marshaled))
	}
}


func (s *Server) broadcast(b []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for ws := range s.conns {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("write err ", err)
				ws.Close()
				s.mu.Lock()
				
				delete(s.conns, ws)
				s.mu.Unlock()
			}
		}(ws)
	}
}

func serveFilesHandler(w http.ResponseWriter, r *http.Request) {
	fs := http.FileServer(http.Dir("./uploads/"))
	fs.ServeHTTP(w, r)
}


func main(){
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				return
			}

			next.ServeHTTP(w, r)
		})
	}
	config, err := appconfig.NewConfig("conf/app_cfg.yml")
	if err != nil {
		fmt.Println("Error loading config: ", err)
	}
	server := NewServer(config)
	http.Handle("/ws", websocket.Handler(server.handleWS))
	http.Handle("/upload", corsMiddleware(http.HandlerFunc(server.handleUpload)))
	http.Handle("/files/", corsMiddleware(http.StripPrefix("/files/", http.FileServer(http.Dir("uploads")))))

	http.Handle("/avatar/", corsMiddleware(http.StripPrefix("/avatar/", http.FileServer(http.Dir("avatars")))))

	port := fmt.Sprintf(":%d", config.Server.Port)
	fmt.Println("Listening on port " + port)
	// http.ListenAndServeTLS(port, "fullchain.pem", "privkey.pem", nil)
	
	http.ListenAndServe(port, nil)
}
