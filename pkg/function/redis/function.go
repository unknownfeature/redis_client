package console

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"redis_client/pkg/cache"
	"strconv"
	"strings"
)

var url = os.Getenv("redis_url")
var promptTemplate = "%s[%s]"
var redisClient *redis.Client

func Execute(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(w, "cannot read body", err)
			return
		}
		var request Request
		err = json.Unmarshal(body, &request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(w, "cannot unmarshal body", err)
			return
		}
		hist := History{Input: request.Command}
		params := strings.Split(request.Command, " ")
		if len(params) < 2 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(w, "unknown command", err)
			return
		}
		theCommand := strings.ToLower(params[0])
		db := request.Db
		oldDb := db
		redisClient = cache.NewClient(db, url)
		if theCommand == "select" {
			changeDb(w, params, hist, oldDb)
		} else {
			executeCommand(w, theCommand, params, hist, db)
			return
		}
		return
	} else if r.Method == http.MethodGet {
		returnPage(w)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	code, err := w.Write([]byte("method not allowed: " + r.Method ))
	log.Println(code, err)
}

func returnPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	f, err := ioutil.ReadFile(os.Getenv("web_template"))
	if err!= nil{
		log.Println("ERROR!!!", err)
	}else {
		log.Println("SUCCESS!!! File exists", string(f))
	}
	t, err := template.ParseFiles(os.Getenv("web_template"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(w, "Unable to load template")
		return
	}
	resp := Response{History: History{}, Prompt: getPrompt(0), Db: 0}

	err = t.Execute(w, resp)
	if err != nil {
		log.Println(err)
	}
	return
}

func getPrompt(db int) string {
	return fmt.Sprintf(promptTemplate, url, strconv.Itoa(db))
}

func executeCommand(w http.ResponseWriter, theCommand string, params []string, hist History, db int) {
	if comm, ok := cache.Commands[theCommand]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		code, err := w.Write([]byte("command not supported: " + theCommand))
		log.Println(code, err)
	} else {
		result := comm(redisClient, context.Background(), params[1:]...)
		hist.Output = result
		respondWithHistory(w, hist, db, db)
	}
}

func changeDb(w http.ResponseWriter, params []string, hist History, db int) {
	newDb, err := strconv.Atoi(params[1])
	if err != nil {
		hist.Output = "(error) ERR invalid DB index"
		w.WriteHeader(http.StatusOK)
		code, err := w.Write([]byte(err.Error()))
		log.Println(code, err)

	} else {
		redisClient = cache.NewClient(newDb, url)
		hist.Output = "OK"
		respondWithHistory(w, hist, newDb, db)
	}
}

func respondWithHistory(w http.ResponseWriter, hist History, db, oldDb int) {
	hist.Input = getPrompt(oldDb) + " " + hist.Input
	resp := Response{History: hist, Prompt: getPrompt(db), Db: db}
	marsh, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		code, err := w.Write([]byte(err.Error()))
		log.Println(code, err)

		return
	}
	w.WriteHeader(http.StatusOK)
	code, err := w.Write(marsh)
	log.Println(code, err)
	return
}
