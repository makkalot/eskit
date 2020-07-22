package common

import (
        "net/http"
	"log"
)

type Gate struct {}

func (gate *Gate) Start() {
        log.Println("Http server listen on : ", 8080)
        http.HandleFunc("/", gate.indexHandler)

        server := &http.Server{Addr: ":8080"}
        go server.ListenAndServe()
}

func (gate *Gate) indexHandler(rw http.ResponseWriter, request *http.Request) {
        rw.WriteHeader(http.StatusOK)
}

