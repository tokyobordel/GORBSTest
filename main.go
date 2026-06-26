package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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

func getCpuFreq() int {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu MHz") {
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
				continue
			}
			valStr := strings.TrimSpace(parts[1])
			freq, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				return 0
			}
			return int(freq)
		}
	}
	return 0
}

func getRamMB() int {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			kbVal, err := strconv.Atoi(fields[1])
			if err != nil {
				return 0
			}
			return kbVal / 1024 // КБ -> МБ
		}
	}
	return 0
}

func telemetryWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		mt, _, _ := conn.ReadMessage()

		infoCpu := getCpuFreq()
		infoMem := getRamMB()

		resp := TelemetryResponse{
			CpuFreq: infoCpu, // Mhz
			Ram:     infoMem, // Mb
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
