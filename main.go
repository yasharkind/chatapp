package main

import (
	"chatapp/base/appconfig"
	controller "chatapp/controllers"
	"chatapp/dto/in"
	"chatapp/dto/in/opcodes"
	"chatapp/dto/out"
	"chatapp/middlewares"
	"chatapp/objects"
	service "chatapp/services"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]string // ws to token

	messageList []*objects.Message

	mu sync.Mutex

	messageService service.MessageService

	userService service.UserService

	userController controller.UserController

	messageController controller.MessageController

	denpaController controller.DenpaController

	config *appconfig.Config
	
	hostname string
}

func NewServer(config *appconfig.Config) *Server {

	usersrv := service.NewUserService(config)
	msgsrv :=  service.NewMessageService(config)
	
	s := &Server{
		config: config,
		conns: make(map[*websocket.Conn]string),
		messageService: msgsrv,
		userService: usersrv,
		userController: *controller.NewUserController(config, usersrv),
		denpaController: *controller.NewDenpaController(config, service.NewUserService(config)),
	}


	s.messageController = *controller.NewMessageController(config, msgsrv, usersrv, s.broadcast);

	if (s.config.Server.TLS) {
		s.hostname = "https://" + config.Server.Host + ":"
	} else {
		s.hostname = "http://" + config.Server.Host + ":"
	}

	s.hostname += strconv.Itoa(s.config.Server.Port)
	println(s.hostname)

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
	fmt.Println("ws from: ", address)
	s.mu.Unlock()

	
	s.readLoop(ws)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {

	file, handler, err := r.FormFile("file")
	os.ReadDir("uploads/")
	token := r.FormValue("token")

	user := s.userService.Validate(token)

	if err != nil {
		fmt.Println("formfile error: ", err)
		w.WriteHeader(400)
		return 
	}

	if user == nil {
		fmt.Println("invalid token")
		w.WriteHeader(400)
		return
	}

	defer file.Close()
	content, err := io.ReadAll(file)
	filenameSplit := strings.Split(handler.Filename, ".")
	filename := strings.Join(filenameSplit[0:len(filenameSplit)-1], "") + time.Now().Format("15_04_05.") + filenameSplit[len(filenameSplit)-1]
	err = os.WriteFile("uploads/" + filename, content, 0644)

	if err != nil {
		fmt.Println("write error: ", err)
	}

	if err != nil {
		fmt.Println("handleUpload error: ", err)
	}

	msg := objects.Message {
		Id: 0,
		Sender: user,
		Content: s.hostname + "/files/" + filename,
		Time:  time.Now().Local(),
	}

	
	s.mu.Lock()
	savedmsg := s.messageService.Save(msg)
	resMsg := dto_out.SendMessage{
			Action: opcodes.SendMessage,
			Id: savedmsg.Id,
			Sender: savedmsg.Sender.Username,
			SenderId: savedmsg.Sender.Id,
			Content: savedmsg.Content,
			Time: savedmsg.Time.Format("15:04:05"),
			Avatar: savedmsg.Sender.Avatar,
		}
	fmt.Println("uplaoded file: ", filename)

	s.mu.Unlock()
	s.broadcast(&resMsg)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	// TODO: move this into a SendMessage http handler then broadcast with opcode from there
	decoder := json.NewDecoder(ws)
	for {
		var msg dto_in.Message
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("read error:", err)
			break
		}

		user := s.userService.Validate(msg.Token)

		if user == nil {
			fmt.Println("user not found")
			continue
		}

		// auth
		switch msg.OpCode{
		case opcodes.Auth:
			s.mu.Lock()
			s.conns[ws] = msg.Token
			s.mu.Unlock()

		case opcodes.SendMessage:
			message := &objects.Message{
				Id: 0,
				Sender: user,
				Time: time.Now().Local(),
				Content: msg.Content,
			}

		
	
			s.mu.Lock()
			savedmsg := s.messageService.Save(*message)
			resMsg := dto_out.SendMessage{
				Action: msg.OpCode,
				Id: savedmsg.Id,
				Sender: savedmsg.Sender.Username,
				SenderId: savedmsg.Sender.Id,
				Content: savedmsg.Content,
				Time: savedmsg.Time.Format("15:04:05"),
				Avatar: savedmsg.Sender.Avatar,
			}
        	
			if strings.HasPrefix(message.Content, ">profile") {
				messageSplit := strings.Split(message.Content, " ")
				if len(messageSplit) == 2 {
					url := messageSplit[1]
					user.Avatar = url
					s.userService.UpdateById(user.Id, *user)
				}
			}
        	
			s.mu.Unlock()

			jsonMsg := fmt.Sprintf("%s-%s: %s", time.Now().Local(), user.Username, msg.Content)
			fmt.Println(jsonMsg)
			s.broadcast(&resMsg)
	
		}	
	}
}


func (s *Server) broadcast(r dto_out.WebSockRes) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for ws := range s.conns {
		go func(ws *websocket.Conn) {
			if user := s.userService.Validate(s.conns[ws]); user == nil {
				s.mu.Lock()
				fmt.Printf("Unauthorized WS: %s != %s\n", s.conns[ws], s.conns[ws])
				ws.Close()

				delete(s.conns, ws)
				s.mu.Unlock()
				return
			}
			switch r.OpCode(){
			case opcodes.SendMessage, opcodes.DeleteMessage, opcodes.EditMessage:
				b, err := json.Marshal(r)
				if err != nil {
					println("Marshal error: ", err)
					return
				}

				if _, err := ws.Write(b); err != nil {
					fmt.Println("write err ", err)
					ws.Close()
					s.mu.Lock()
					
					delete(s.conns, ws)
					s.mu.Unlock()
				}
			default:
				println("Unknow OP Code ", r.OpCode())
			}

		}(ws)
	}
}


func main(){
	config, err := appconfig.NewConfig("conf/app_cfg.yml")
	if err != nil {
		fmt.Println("error loading config: ", err)
	}
	server := NewServer(config)
	http.Handle("/ws", websocket.Handler(server.handleWS))
	http.Handle("/upload", middlewares.CorsMiddleware(http.HandlerFunc(server.handleUpload)))
	http.Handle("/files/", middlewares.CorsMiddleware(http.StripPrefix("/files/", http.FileServer(http.Dir("uploads")))))


	port := fmt.Sprintf(":%d", server.config.Server.Port)
	fmt.Println("Listening on port " + port)

	if server.config.Server.TLS {
		http.ListenAndServeTLS(port, "fullchain.pem", "privkey.pem", nil)
	} else {
		http.ListenAndServe(port, nil)
	}
}
