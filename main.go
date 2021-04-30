package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var Tbl map[string]int
var err error
var conn redis.Conn
var indexTemplate *template.Template
var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

var Clients = make(map[*websocket.Conn]bool)
var pool = NewPool()

type response1 struct {
	Token  string
	Value  string
	Expiry time.Duration
	Keys   []string
}

func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func StartSession(key string) {
	Tbl[key] = 1
	time.Sleep(10 * time.Second)
	defer fmt.Printf("Session %s ended\n", key)
}

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Worker %d commenced\n", id)
	start := time.Now()
	time.Sleep(time.Second * 2)
	end := start.Add(time.Duration(time.Second * 2))
	fmt.Println(start)
	fmt.Println(end)
	fmt.Printf("Worker %d finished\n", id)
}

func Basic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token := randToken()
	duration := 3

	res, err := Setex(conn, token, "Good Morning", duration)
	if err != nil {
		log.Fatalln(err)
	}

	keys, err := Keys(conn)
	if err != nil {
		log.Fatalln(err)
	}
	result := &response1{token, res, time.Duration(duration), keys}
	data, _ := json.Marshal(result)
	w.Write(data)
}

func Check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key := chi.URLParam(r, "roomID")
	res, err := Get(conn, key)
	if err != nil {
		log.Println(err)
		log.Println(res)
		http.Error(w, "Key Error", http.StatusInternalServerError)
	}
	data, err := json.Marshal(map[string]string{key: res})
	if err != nil {
		log.Println(err)
		panic(err)
	}
	w.Write(data)
}

func Homepage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	indexTemplate.Execute(w, "Nothing")
}

func Reader(ws *websocket.Conn) {
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
		}
		fmt.Println(string(msg))
	}
}

func Writer(ws *websocket.Conn) {
	for {
		err = ws.WriteMessage(websocket.TextMessage, []byte("Good Night!"))
		if err != nil {
			log.Println(err)
		}
	}
}

func Chat(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Websocket upgrade failed", http.StatusMethodNotAllowed)
		return
	}

	client := &Client{Conn: ws, Pool: pool}
	pool.Register <- client
	client.Read()
}

func NewChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key := chi.URLParam(r, "id")

	// Check the redis cache for the key
	res, err := Get(conn, key)
	if err != nil {
		log.Println(err)
		http.Error(w, "Key Error: Key expired or does not exist!", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(map[string]string{key: res})
	if err != nil {
		log.Println(err)
		panic(err)
	}

	w.Write(data)
}

func GenerateURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := uuid.NewString()
	res, err := Setex(conn, id, "Activated", 3)
	if err != nil {
		http.Error(w, "Redis SETEX error!", http.StatusInternalServerError)
	}
	fmt.Printf("GenerateURL : %s\n", res)
	w.Write([]byte(id))
}

func main() {
	conn, err = redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
	defer conn.Close()

	// Setup connection pool
	go pool.Start()

	// Templates
	indexTemplate, err = template.ParseFiles("./templates/index.html")
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	const PORT = ":5000"
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.URLFormat)

	r.Get("/json", Basic)

	r.Get("/{roomID}", Check)

	r.Get("/", Homepage)
	r.Get("/ws", Chat)
	r.Get("/chat/{id}", NewChat)
	r.Get("/generate", GenerateURL)
	fmt.Println("Running on port " + PORT)
	http.ListenAndServe(PORT, r)
}
