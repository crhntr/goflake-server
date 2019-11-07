package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/bstick12/goflake"
)

func main() {
	var generator = goflake.GoFlakeInstanceUsingUnique("D01Z01")

	router := http.NewServeMux()
	router.HandleFunc("/ids", ensureMethod(http.MethodGet, idList(generator)))
	router.HandleFunc("/health", ensureMethod(http.MethodGet, healthCheck))
	log.Fatal(http.ListenAndServe(":8080", router))
}

type base64UUIDGetter interface {
	GetBase64UUID() string
}

//http.HandlerFunc: "GET" "/ids"
func idList(generator base64UUIDGetter) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		count, ok := countParam(res, req)
		if !ok {
			return
		}

		res.WriteHeader(http.StatusOK)

		ids := make([]string, count)
		for index := range ids {
			ids[index] = generator.GetBase64UUID()
		}
		json.NewEncoder(res).Encode(ids)
	}
}

const maxCount = 1000

var countErrorMessage = fmt.Sprintf("count query parameter must be a valid positive integer less than or equal to %d", maxCount)

// /?count=int
func countParam(res http.ResponseWriter, req *http.Request) (int, bool) {
	count := 1
	countString := req.URL.Query().Get("count")
	if countString != "" {
		var err error
		count, err = strconv.Atoi(countString)
		if err != nil || count < 0 || count > maxCount {
			respondWithErrorMessage(res, req, countErrorMessage, http.StatusBadRequest)
			return 0, false
		}
	}
	return count, true
}

//http.HandlerFunc: "GET" "/health-check"
func healthCheck(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("cache-control", "no-cache")
	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("I'm okay.\n"))
}

func respondWithErrorMessage(res http.ResponseWriter, req *http.Request, message string, status int) {
	res.Header().Set("Content-Type", "application/json; charset=UTF-8")
	res.WriteHeader(status)
	json.NewEncoder(res).Encode(struct {
		Error string `json:"error"`
	}{Error: message})
}

func ensureMethod(method string, fn http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			respondWithErrorMessage(res, req,
				"this method is not supported",
				http.StatusMethodNotAllowed)
			return
		}
		fn(res, req)
	}
}
