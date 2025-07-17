package main

import (
	"CommentsEngine/redisc"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"net/http"
	"os"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

var (
	httpHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "comments_engine_http_hits_total",
		Help: "The total number of hits on a given route",
	}, []string{"route", "method"})
	commentsReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "comments_engine_comments_received_total",
		Help: "The total number of comments received",
	})
	errorsSent = promauto.NewCounter(prometheus.CounterOpts{
		Name: "comments_interactor_errors_sent_total",
		Help: "The total number of errors returned",
	})
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	httpHits.With(prometheus.Labels{"route": "/", "method": r.Method}).Inc()
	io.WriteString(w, "CommentsEngine answering here !\n")
}

func sendError(w http.ResponseWriter, err error, additionalInfo string) {
	errorsSent.Inc()
	w.WriteHeader(http.StatusInternalServerError)
	resp := make(map[string]string)
	resp["message"] = fmt.Sprint(err)
	resp["additionalInfo"] = additionalInfo
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func sendMessage(w http.ResponseWriter, message string) {
	commentsReceived.Inc()
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func getLatestComments(w http.ResponseWriter, r *http.Request) {
	httpHits.With(prometheus.Labels{"route": "/latest", "method": r.Method}).Inc()
	w.Header().Set("Content-Type", "application/json")
	messages, err := redisc.GetLatestComments()
	if err == nil {
		io.WriteString(w, messages)
	} else {
		sendError(w, err, "Maybe there is a problem with redis backend ?")
	}
}

type Comment struct {
	Comment string `json:"comment"`
}

func sendComment(w http.ResponseWriter, r *http.Request) {
	httpHits.With(prometheus.Labels{"route": "/comment", "method": r.Method}).Inc()
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		sendError(w, errors.New("POST not used"), "HTTP method was not POST")
		return
	}
	comment := &Comment{}
	err := json.NewDecoder(r.Body).Decode(comment)
	if err != nil {
		sendError(w, err, "The data sent was likely incorrect")
		return
	}
	_, err = redisc.StoreComment(comment.Comment)
	if err != nil {
		sendError(w, err, "Problem with redis maybe ?")
		return
	}
	sendMessage(w, "stored !")
}

func main() {
	http.HandleFunc("/", getRoot)
	http.HandleFunc("/latest", getLatestComments)
	http.HandleFunc("/comment", sendComment)
	http.Handle("/metrics", promhttp.Handler())

	redisc.Client = redis.NewClient(&redis.Options{
		Addr:     getenv("REDIS_ENDPOINT", "localhost:6379"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	listeningHostPort := getenv("HOST", "127.0.0.1") + ":" + getenv("PORT", "8000")
	fmt.Printf("Starting the server on %s\n", listeningHostPort)
	err := http.ListenAndServe(listeningHostPort, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
