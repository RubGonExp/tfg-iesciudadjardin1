package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Todo struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Updated       time.Time `json:"updated"`
	Completed     time.Time `json:"completed"`
	Complete      bool      `json:"complete"`
	completedNull sql.NullTime
}

// JSON transforma el contenido de una tarea en json.
func (t Todo) JSON() (string, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("could not marshal json for response: %s", err)
	}

	return string(bytes), nil
}

// JSONBytes transforma el contenido de una tarea en json como una matriz de bytes.
func (t Todo) JSONBytes() ([]byte, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return []byte{}, fmt.Errorf("could not marshal json for response: %s", err)
	}

	return bytes, nil
}

// Key devuelve el id en forma de cadena.
func (t Todo) Key() string {
	return strconv.Itoa(t.ID)
}

type Todos []Todo

// JSON transforma el contenido de una porcion de todos en json.
func (t Todos) JSON() (string, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("could not marshal json for response: %s", err)
	}

	return string(bytes), nil
}

// JSONBytes transforma el contenido de una porcion de todos en json como una matriz de bytes.
func (t Todos) JSONBytes() ([]byte, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return []byte{}, fmt.Errorf("could not marshal json for response: %s", err)
	}

	return bytes, nil
}
