package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type TelemetryResponse struct {
	CpuFreq int `json:"cpu_freq"`
	Ram     int `json:"ram"`
}

type EchoRequest struct {
	Text string `json:"text"`
}

type EchoResponse struct {
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	// Разрешаем соединения с любых источников (в продакшене лучше ограничить)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func telemetryWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		mt, _, _ := conn.ReadMessage()

		infoCpu, _ := cpu.Info()
		infoMem, _ := mem.VirtualMemory()

		resp := TelemetryResponse{
			CpuFreq: int(infoCpu[0].Mhz),                // Mhz
			Ram:     int(infoMem.Total / (1024 * 1024)), // Mb
		}

		respBytes, _ := json.Marshal(resp)

		conn.WriteMessage(mt, respBytes)

	}
}

func echoWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		mt, message, _ := conn.ReadMessage()

		var req EchoRequest
		json.Unmarshal(message, &req)

		resp := EchoResponse{
			Text:      req.Text,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		respBytes, _ := json.Marshal(resp)

		conn.WriteMessage(mt, respBytes)

	}

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/telemetry", telemetryWebSocket).Methods("GET")
	r.HandleFunc("/echo", echoWebSocket).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	port := ":8080"
	log.Fatal(http.ListenAndServe(port, r))
}
