package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

const (
	web_socket_password = "1qaz2wsx#EDCXZASWQ@!!"
)

type ClientManager struct {
	Clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	id     string
	socket *websocket.Conn
	send   chan []byte
	check  bool
	times  int
}
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var Manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	Clients:    make(map[*Client]bool),
}

var LoginList = make(map[string]bool)


var upGrader = &websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}



func (Manager *ClientManager) start() {
	for {
		select {
		case conn := <-Manager.register:
			Manager.Clients[conn] = true

		case conn := <-Manager.unregister:
			if _, ok := Manager.Clients[conn]; ok {
				close(conn.send)
				delete(Manager.Clients, conn)
			}
		case message := <-Manager.broadcast:
			for conn := range Manager.Clients {
				select {
				case conn.send <- message:
				default:
					// close(conn.send)
					// delete(manager.clients, conn)
				}
			}
		}
	}
}

// 回傳Websocker的服務
func Get_manager() *ClientManager {
	return &Manager
}


// 傳送Web監控的資料
func Send_ws_data(data []byte) {
	if len(Manager.Clients) != 0 {
		for conn := range Manager.Clients {
			// conn.socket.WriteMessage(websocket.CloseMessage, []byte{})
			if conn.check {
				conn.send <- data
			}

		}
	}
}

// 接收從client端傳送的訊息
func (c *Client) read() {
	defer func() {
		Manager.unregister <- c
		c.socket.Close()
	}()
	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			c.check = false
			Manager.unregister <- c
			break
		}
		password := string(message[:])
		// fmt.Printf(password, "\n")
		if web_socket_password != password {
			Manager.unregister <- c
			break
		} else {
			c.check = true
		}

	}
}

// 傳送訊息給client端
func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// websocker的服務動作
func Start_websocket() {
	// 讓websocker的服務  自己跑一個thread
	go Manager.start()

}


// 針對websocket的服務 開啟 /ws的url 及 1234此端口
func NewWebClient(res http.ResponseWriter, req *http.Request){
	conn, error := upGrader.Upgrade(res, req, nil)
	if error != nil {
		http.NotFound(res, req)
		return
	}
	u := uuid.NewV4()
	obj := u.String()
	client := &Client{id: obj, socket: conn, send: make(chan []byte), check: false, times: 0}

	Manager.register <- client

	go client.read()
	go client.write()
}