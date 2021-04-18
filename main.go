package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gomodule/redigo/redis"
)

var Tbl map[string]int
var conn redis.Conn

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
		panic(err)
	}

	keys, err := Keys(conn)
	if err != nil {
		panic(err)
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
		panic(err)
	}
	data, err := json.Marshal(map[string]string{key: res})
	if err != nil {
		panic(err)
	}
	w.Write(data)
}

func main() {
	var err error
	conn, err = redis.Dial("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	const PORT = ":3000"
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.URLFormat)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	r.Get("/json", Basic)

	r.Get("/{roomID}", Check)

	fmt.Println("Running on port " + PORT)
	http.ListenAndServe(PORT, r)
}
