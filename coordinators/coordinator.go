package coordinators

import (
	"bytes"
	"encoding/gob"
	"first-distributed-system/dto"
	"first-distributed-system/qutils"
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

const THRESHOLD_TIME float64 = 5

var (
	eventSensorValueReceived = "sensor value received"
)

type SensorsListener struct {
	sensors map[string]chan bool
	ea      EventAggregator
}

type printListener struct{}

func (l *printListener) Handle(data EventData) {
	fmt.Printf("Sensor: %v. Value: %v. Timestamp: %v\n", data.Name, data.Value, data.Timestamp)
}

type throttleListener struct {
	lastValueTimestamps map[string]time.Time
}

func newThrottleListener() *throttleListener {
	l := new(throttleListener)
	l.lastValueTimestamps = make(map[string]time.Time)

	return l
}

func (l *throttleListener) Handle(data EventData) {
	// if last value is present AND time passed is < THRESHOLD_TIME - skip value
	// else - push the value to the persistent queue
	if t, ok := l.lastValueTimestamps[data.Name]; !ok || data.Timestamp.Sub(t).Seconds() > THRESHOLD_TIME {
		qutils.DeclareQueue("amq.direct", qutils.QueuePersistentValues, true, []string{"sensor.value"})
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		enc.Encode(data)
		l.lastValueTimestamps[data.Name] = data.Timestamp
		qutils.PublishToDirect("amq.direct", "sensor.value", buf.Bytes())
	}
}

func NewSensorsListener() *SensorsListener {
	l := SensorsListener{
		sensors: make(map[string]chan bool),
		ea:      NewEventAggregator(),
	}

	l.ea.AddListener(eventSensorValueReceived, new(printListener))
	l.ea.AddListener(eventSensorValueReceived, newThrottleListener())

	return &l
}

func (l *SensorsListener) ListenForSensors() {
	q := qutils.DeclareQueue(qutils.SensorsListExchange, "", true, []string{""})
	qutils.Consume(q.Name, func(msg amqp.Delivery) {
		sensorName := string(msg.Body)
		if v := l.sensors[sensorName]; v == nil {
			l.sensors[sensorName] = make(chan bool)
			go qutils.PublishToFanout(qutils.WebappSourcesExchange, []byte(sensorName))
			go l.listenForSensor(sensorName)
		}
	})
}

func (l *SensorsListener) ListenForWebAppDiscovery() {
	qutils.DeclareFanoutExchange(qutils.WebappDiscoverExchange)
	q := qutils.DeclareQueue(qutils.WebappDiscoverExchange, "", true, []string{""})

	qutils.Consume(q.Name, func(msg amqp.Delivery) {
		for sensor := range l.sensors {
			qutils.PublishToFanout(qutils.WebappSourcesExchange, []byte(sensor))
		}
	})
}

func (l *SensorsListener) ListenForSensorsDisabled() {
	qutils.DeclareFanoutExchange(qutils.SensorsDisabledExchange)
	q := qutils.DeclareQueue(qutils.SensorsDisabledExchange, "", true, []string{""})

	fmt.Println("start listening for sensors disabling")

	qutils.Consume(q.Name, func(msg amqp.Delivery) {
		sensor := string(msg.Body)
		for s, _ := range l.sensors {
			if s == sensor {
				delete(l.sensors, s)
			}
		}
	})
}

func (l *SensorsListener) DiscoverSensors() {
	qutils.PublishToFanout(qutils.SensorsDiscoverExchange, []byte("hello"))
}

func (l *SensorsListener) listenForSensor(sensorName string) {
	q := qutils.DeclareQueue("", sensorName, true, []string{})
	qutils.Consume(q.Name, func(msg amqp.Delivery) {
		buf := bytes.NewReader(msg.Body)
		d := gob.NewDecoder(buf)
		value := new(dto.SensorMessage)
		d.Decode(value)

		l.ea.FireEvent(eventSensorValueReceived, EventData{
			Name:      value.Name,
			Value:     value.Value,
			Timestamp: value.Timestamp,
		})
	})
}
