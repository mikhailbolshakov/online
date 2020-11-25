package server

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)


type httpServer struct {
	server *http.Server
	wsUpgrader *websocket.Upgrader
	roomService *RoomHttpService
	webSocketService *WebSocketService
}

func (ws *WsServer) init() *httpServer {

	router := mux.NewRouter()

	srv := &http.Server{
		Addr: ws.port,
		Handler: router,
		// TODO: take from env
		WriteTimeout: time.Hour,
		ReadTimeout: time.Hour,
	}

	server := &httpServer{
		server:      srv,
		wsUpgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		roomService: &RoomHttpService{
			ws: ws,
		},
		webSocketService: &WebSocketService{
			ws: ws,
		},
	}

	server.webSocketService.setRouting(router)
	server.roomService.setRouting(router)

	return server
}

func (ws *WsServer) listenAndServe() {

	ws.httpServer = ws.init()
	log.Fatal(ws.httpServer.server.ListenAndServe())

}

func (http *httpServer) respondWithError(w http.ResponseWriter, code int, message string) {
	http.respondWithJSON(w, code, map[string]string{"error": message})
}

func (http *httpServer) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

