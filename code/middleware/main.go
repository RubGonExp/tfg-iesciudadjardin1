package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var storage Storage

func main() {
	user := os.Getenv("todo_user")
	pass := os.Getenv("todo_pass")
	host := os.Getenv("todo_host")
	name := os.Getenv("todo_name")
	redisHost := os.Getenv("REDISHOST")
	redisPort := os.Getenv("REDISPORT")
	port := os.Getenv("PORT")

	fmt.Printf("Port: %s\n", port)

	if err := storage.Init(user, pass, host, name, redisHost, redisPort, true); err != nil {
		panic(err)
	}
	defer storage.sqlstorage.Close()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/healthz", healthHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/healthz", healthHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/todo", listHandler).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/api/v1/todo", createHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/todo/{id}", readHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/todo/{id}", deleteHandler).Methods(http.MethodDelete)
	router.HandleFunc("/api/v1/todo/{id}", updateHandler).Methods(http.MethodPost, http.MethodPut)

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}

// CORSRouterDecorator aplica cabeceras CORS a un mux.Router
type CORSRouterDecorator struct {
	R *mux.Router
}

// ServeHTTP envuelve el servidor HTTP habilitando las cabeceras CORS.
func (c *CORSRouterDecorator) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers", "Accept, Accept-Language, Content-Type, YourOwnHeader")
	}
	// Stop aqui si se trata de una solicitud de OPCIONES PREVISTAS
	if req.Method == "OPTIONS" {
		return
	}

	c.R.ServeHTTP(rw, req)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
	return
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	ts, err := storage.List()
	if err != nil {
		writeErrorMsg(w, err)
		return
	}

	writeJSON(w, ts, http.StatusOK)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	_, err := strconv.Atoi(id)
	if err != nil {
		msg := Message{"invalid! id must be integer", fmt.Sprintf("todo id: %s", id)}
		writeJSON(w, msg, http.StatusInternalServerError)
		return
	}

	t, err := storage.Read(id)
	if err != nil {

		if strings.Contains(err.Error(), "Rows are closed") {
			msg := Message{"todo not found", fmt.Sprintf("todo id: %s", id)}
			writeJSON(w, msg, http.StatusNotFound)
			return
		}

		writeErrorMsg(w, err)
		return
	}

	writeJSON(w, t, http.StatusOK)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	t := Todo{}
	t.Title = r.FormValue("title")

	if len(r.FormValue("complete")) > 0 && r.FormValue("complete") != "false" {
		t.Complete = true
	}

	t, err := storage.Create(t)
	if err != nil {
		writeErrorMsg(w, err)
		return
	}

	writeJSON(w, t, http.StatusCreated)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	t := Todo{}
	id := mux.Vars(r)["id"]
	t.ID, err = strconv.Atoi(id)
	if err != nil {
		writeErrorMsg(w, err)
		return
	}

	t.Title = r.FormValue("title")

	if len(r.FormValue("complete")) > 0 && r.FormValue("complete") != "false" {
		t.Complete = true
	}

	if err = storage.Update(t); err != nil {
		writeErrorMsg(w, err)
		return
	}

	writeJSON(w, t, http.StatusOK)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	_, err := strconv.Atoi(id)
	if err != nil {
		msg := Message{"invalid! id must be integer", fmt.Sprintf("todo id: %s", id)}
		writeJSON(w, msg, http.StatusInternalServerError)
		return
	}

	if err := storage.Delete(id); err != nil {
		writeErrorMsg(w, err)
		return
	}
	msg := Message{"todo deleted", fmt.Sprintf("todo id: %s", id)}

	writeJSON(w, msg, http.StatusNoContent)
}

// JSONProducer es una interfaz que escupe una version en cadena JSON de si misma
type JSONProducer interface {
	JSON() (string, error)
	JSONBytes() ([]byte, error)
}

func writeJSON(w http.ResponseWriter, j JSONProducer, status int) {
	json, err := j.JSON()
	if err != nil {
		writeErrorMsg(w, err)
		return
	}
	writeResponse(w, status, json)
	return
}

func writeErrorMsg(w http.ResponseWriter, err error) {
	s := fmt.Sprintf("{\"error\":\"%s\"}", err)
	writeResponse(w, http.StatusInternalServerError, s)
	return
}

func writeResponse(w http.ResponseWriter, status int, msg string) {
	if status != http.StatusOK {
		weblog(fmt.Sprintf(msg))
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")
	w.WriteHeader(status)
	w.Write([]byte(msg))

	return
}

func weblog(msg string) {
	log.Printf("Webserver : %s", msg)
}

// Message es una estructura para comunicar datos adicionales al consumidor de la API.
type Message struct {
	Text    string `json:"text"`
	Details string `json:"details"`
}

// JSON transforma el contenido de una tarea en json.
func (m Message) JSON() (string, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("could not marshal json for response: %s", err)
	}

	return string(bytes), nil
}

// JSONBytes transforma el contenido de una tarea en json como una matriz de bytes.
func (m Message) JSONBytes() ([]byte, error) {
	bytes, err := json.Marshal(m)
	if err != nil {
		return []byte{}, fmt.Errorf("could not marshal json for response: %s", err)
	}

	return bytes, nil
}
