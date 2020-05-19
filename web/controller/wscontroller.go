package controller

import (
	"bytes"
	"encoding/gob"
	"first-distributed-system/dto"
	"first-distributed-system/qutils"
	"first-distributed-system/web/model"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"sync"
)

type wsController struct {
	sockets  []*websocket.Conn
	mutex    sync.Mutex
	upgrader websocket.Upgrader
}

type message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func newWSController() *wsController {
	wsc := new(wsController)
	wsc.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	go wsc.listenForSources()
	go wsc.listenForMessages()
	go wsc.listenForSensorsDisabling()

	return wsc
}

func (wsc *wsController) handleMessage(w http.ResponseWriter, r *http.Request) {
	socket, _ := wsc.upgrader.Upgrade(w, r, nil)
	wsc.addSocket(socket)

	go wsc.listenForDiscoveryRequests(socket)
}

func (wsc *wsController) addSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	wsc.sockets = append(wsc.sockets, socket)
	wsc.mutex.Unlock()
}

func (wsc *wsController) removeSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	socket.Close()

	for i := range wsc.sockets {
		if wsc.sockets[i] == socket {
			wsc.sockets = append(wsc.sockets[:i], wsc.sockets[i+1:]...)
		}
	}

	wsc.mutex.Unlock()
}

func (wsc *wsController) listenForDiscoveryRequests(socket *websocket.Conn) {
	for {
		msg := message{}
		err := socket.ReadJSON(&msg)

		if err != nil {
			wsc.removeSocket(socket)
			break
		}

		if msg.Type == "discover" {
			qutils.DeclareFanoutExchange(qutils.WebappDiscoverExchange)
			qutils.PublishToFanout(qutils.WebappDiscoverExchange, []byte{})
		}

	}
}

func (wsc *wsController) listenForSources() {
	qutils.DeclareFanoutExchange(qutils.WebappSourcesExchange)
	q := qutils.DeclareQueue(qutils.WebappSourcesExchange, "", true, []string{""})

	qutils.Consume(q.Name, func(msg amqp.Delivery) {
		sensor := model.Sensor{
			Name: string(msg.Body),
		}
		log.Println("got new sensor", sensor)

		wsc.sendMessage(message{
			Type: "source",
			Data: sensor,
		})
	})
}

func (wsc *wsController) listenForSensorsDisabling() {
	qutils.DeclareFanoutExchange(qutils.SensorsDisabledExchange)
	q := qutils.DeclareQueue(qutils.SensorsDisabledExchange, "", true, []string{""})

	qutils.Consume(q.Name, func(msg amqp.Delivery) {
		sensor := model.Sensor{
			Name: string(msg.Body),
		}
		log.Println("sensor disabled", sensor)

		wsc.sendMessage(message{
			Type: "disabled",
			Data: sensor,
		})
	})
}

func (wsc *wsController) sendMessage(m message) {
	var socketsToRemove []*websocket.Conn

	for _, s := range wsc.sockets {
		err := s.WriteJSON(m)

		if err != nil {
			socketsToRemove = append(socketsToRemove, s)
		}
	}

	for _, s := range socketsToRemove {
		wsc.removeSocket(s)
	}
}

func (wsc *wsController) listenForMessages() {
	q := qutils.DeclareQueue("amq.direct", "webappValues", true, []string{"sensor.value"})
	qutils.Consume(q.Name, func(msg amqp.Delivery) {
		sm := dto.SensorMessage{}
		buf := bytes.NewReader(msg.Body)
		dec := gob.NewDecoder(buf)
		dec.Decode(&sm)

		wsc.sendMessage(message{
			Type: "reading",
			Data: sm,
		})
	})
}
