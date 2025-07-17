package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

type Comment struct {
	Comment string `json:"comment"`
}

type EngineError struct {
	Message        string `json:"message"`
	AdditionalInfo string `json:"additionalInfo"`
}

var comments_engine_endpoint string
var (
	httpHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "comments_interactor_http_hits_total",
		Help: "The total number of hits on a given route",
	}, []string{"route", "method"})
	commentsReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "comments_interactor_comments_received_total",
		Help: "The total number of comments received (successfully posted or not)",
	})
	commentsPosted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "comments_interactor_comments_posted_total",
		Help: "The total number of comments posted (successfully posted)",
	})
)

func sendComment(w http.ResponseWriter, r *http.Request) {
	commentsReceived.Inc()
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "<h1>error</h1><p>ParseForm() err: %v</p>", err)
		return
	}
	var comment Comment
	comment.Comment = r.FormValue("comment")
	if comment.Comment == "" {
		fmt.Fprintf(w, "<h1>error</h1><p>Empty comment or malformed</p>")
		return
	}
	obj, err := json.Marshal(comment)
	if err != nil {
		fmt.Fprintf(w, "<h1>error</h1><p>Empty comment or malformed</p>")
		return
	}
	_, err = http.Post("http://"+comments_engine_endpoint+"/comment", "application/json", bytes.NewReader(obj))
	if err == nil {
		fmt.Fprintf(w, "<h1>Success !</h1><p>Comment posted</p>")
		commentsPosted.Inc()
	} else {
		fmt.Fprintf(w, "<h1>Error</h1><p>Error while trying to send a comment: %s", err)
	}
}

func showDashboard(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "<h1>Latest comments</h1>")
	io.WriteString(w, "<p>")
	resp, err := http.Get("http://" + comments_engine_endpoint + "/latest")
	if err != nil {
		log.Printf("Could not HTTP get on CommentsEngine: %s", err)
		io.WriteString(w, fmt.Sprintf("Error contacting the backend, %s", err))
	} else {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		var comments []string
		err = json.Unmarshal(body, &comments)
		if err != nil {
			var engineError EngineError
			err = json.Unmarshal(body, &engineError)
			if err != nil {
				log.Printf("CommentsEngine returned unexpected value: %s", err)
				io.WriteString(w, fmt.Sprintf("CommentsEngine returned badly formated values, %s", err))
			} else {
				io.WriteString(w, fmt.Sprintf("CommentsEngine returned an error: %s<br/>Additional info: %s", engineError.Message, engineError.AdditionalInfo))
			}
		} else {
			for i, comment := range comments {
				io.WriteString(w, "<div><h3>comment "+strconv.Itoa(i)+"</h3><p>"+comment+"</p></div>")
			}
		}
	}
	io.WriteString(w, "</p>")
	io.WriteString(w, "<h1>Send comment</h1>")
	io.WriteString(w, "<form method=\"post\"><textarea name=\"comment\" autofocus=\"true\" placeholder=\"your comment ...\" rows=\"5\" cols=\"80\"></textarea><br/><input type=\"submit\" value=\"send comment\"></form>")
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	httpHits.With(prometheus.Labels{"route": "/", "method": r.Method}).Inc()
	io.WriteString(w, "<!doctype html><html><head><title>CommentsInteractor</title></head><body>")
	if r.Method == http.MethodPost {
		sendComment(w, r)
	}
	showDashboard(w, r)
	io.WriteString(w, "</body></html>")
}

func sendError(w http.ResponseWriter, err error, additionalInfo string) {
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

func main() {
	http.HandleFunc("/", getRoot)
	prometheus.Unregister(prometheus.NewGoCollector())
	http.Handle("/metrics", promhttp.Handler())

	listeningHostPort := getenv("HOST", "127.0.0.1") + ":" + getenv("PORT", "9000")
	comments_engine_endpoint = getenv("COMMENTS_ENGINE_ENDPOINT", "127.0.0.1:8000")
	fmt.Printf("Starting the server on %s\n", listeningHostPort)
	err := http.ListenAndServe(listeningHostPort, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
