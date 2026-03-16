package main

import (
	"chatapp/base/appconfig"
	"chatapp/dto/in"
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

	config *appconfig.Config
	
	hostname string
}

func NewServer(config *appconfig.Config) *Server {
	
	s := &Server{
		config: config,
		conns: make(map[*websocket.Conn]string),
		messageService: service.NewMessageService(config),
		userService: service.NewUserService(config),
	}

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

	user := middlewares.Validate(token, s.config.Secret, s.userService)

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
	resMsg := dto_out.Message{
			Id: savedmsg.Id,
			Sender: savedmsg.Sender.Username,
			Content: savedmsg.Content,
			Time: savedmsg.Time.Format("15:04:05"),
			Avatar: savedmsg.Sender.Avatar,
		}
	fmt.Println("uplaoded file: ", filename)
	marshaled, err := json.Marshal(resMsg)
	if err != nil {
		fmt.Println("marshal err: ", err)	
	}

	s.mu.Unlock()
	s.broadcast([]byte(marshaled))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) readLoop(ws *websocket.Conn) {
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

		user := middlewares.Validate(msg.Token, s.config.Secret, s.userService)

		if user == nil {
			fmt.Println("user not found")
			continue
		}

		// auth
		println(msg.OpCode)
		if msg.OpCode == 0 {
			s.mu.Lock()
			s.conns[ws] = msg.Token
			println(msg.Token)
			s.mu.Unlock()
			continue
		}
		
		//jsonMsg := fmt.Sprintf("%s-%s: %s", time.Now().Local(), user.Username, msg.Content)
		//fmt.Println(jsonMsg)
		message := &objects.Message{
			Id: 0,
			Sender: user,
			Time: time.Now().Local(),
			Content: msg.Content,
		}

		
	
		s.mu.Lock()
		savedmsg := s.messageService.Save(*message)
		resMsg := dto_out.Message{
			Id: savedmsg.Id,
			Sender: savedmsg.Sender.Username,
			Content: savedmsg.Content,
			Time: savedmsg.Time.Format("15:04:05"),
			Avatar: savedmsg.Sender.Avatar,
		}
		marshaled, err := json.Marshal(resMsg)
		if err != nil {
			fmt.Println("Error marshaling: ", err)
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

		s.broadcast([]byte(marshaled))
	}
}


func (s *Server) broadcast(b []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for ws := range s.conns {
		go func(ws *websocket.Conn) {
			if user := middlewares.Validate(s.conns[ws], s.config.Secret, s.userService); user == nil {
				s.mu.Lock()
				fmt.Printf("Unauthorized WS: %s != %s\n", s.conns[ws], s.conns[ws])
				ws.Close()

				delete(s.conns, ws)
				s.mu.Unlock()
				return
			}
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

func (s *Server) handleMessageHistory(w http.ResponseWriter, r *http.Request) {
	token := strings.Join(r.Header["Authorization"], "")
	if (token == "") {
		fmt.Println("token is empty")
		return
	}
	user := middlewares.Validate(token, s.config.Secret, s.userService)
	if user == nil {
		fmt.Println("user not found")
		return
	}
	s.mu.Lock()
	messageList := s.messageService.FindFromEnd(s.config.Server.MessageLimit)
	
	var messageDtos []dto_out.Message
	for _, message := range messageList {
		messageDtos = append(messageDtos, dto_out.Message{
			Id: message.Id,
			Sender: message.Sender.Username,
			Content: message.Content,
			Avatar: message.Sender.Avatar,
			Time: message.Time.Local().Format("15:04:05"),
		})
	}
	dto := dto_out.Auth{
		Token: token,
		Messages: messageDtos,
	}
	marshaled, err := json.Marshal(dto)
	if err != nil {
		fmt.Println("marshal error: ", err)
	}
	if _, err := w.Write(marshaled); err != nil {
		fmt.Println("write error: ", err)
	}
	s.mu.Unlock()
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request){
	var user_dto *dto_in.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user_dto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("decode error: ", err)
		return
	}
	token, err := s.userService.JWTByUsernameAndPassword(user_dto.Username, user_dto.Password)
	if err != nil {
		fmt.Println("jwt error: ", err)
		w.WriteHeader(403)
	}
	w.Write([]byte(token))
}

func main(){
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				return
			}

			next.ServeHTTP(w, r)
		})
	}
	config, err := appconfig.NewConfig("conf/app_cfg.yml")
	if err != nil {
		fmt.Println("error loading config: ", err)
	}
	server := NewServer(config)
	http.Handle("/ws", websocket.Handler(server.handleWS))
	http.Handle("/upload", corsMiddleware(http.HandlerFunc(server.handleUpload)))
	http.Handle("/files/", corsMiddleware(http.StripPrefix("/files/", http.FileServer(http.Dir("uploads")))))

	http.Handle("/avatar/", corsMiddleware(http.StripPrefix("/avatar/", http.FileServer(http.Dir("avatars")))))

	http.Handle("/auth", corsMiddleware(http.HandlerFunc(server.handleMessageHistory)))
	http.Handle("/login", corsMiddleware(http.HandlerFunc(server.handleLogin)))

	port := fmt.Sprintf(":%d", server.config.Server.Port)
	fmt.Println("Listening on port " + port)

	if server.config.Server.TLS {
		http.ListenAndServeTLS(port, "fullchain.pem", "privkey.pem", nil)
	} else {
		http.ListenAndServe(port, nil)
	}
}
