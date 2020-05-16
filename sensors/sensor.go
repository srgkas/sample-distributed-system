package main

import (
	"bytes"
	"encoding/gob"
	"first-distributed-system/dto"
	"first-distributed-system/qutils"
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"strconv"
	"time"
)

var name = flag.String("name", "sensor", "name of the sensor")
var freq = flag.Uint("freq", 5, "update frequency in cycles/sec")
var max = flag.Float64("max", 5., "max value fo generated readings")
var min = flag.Float64("min", 1., "min value fo generated readings")
var stepSize = flag.Float64("step", 0.1, "max allowable change per measurement")

var randomizer = rand.New(rand.NewSource(time.Now().UnixNano()))

var valueRange, nom, value float64

func main() {
	flag.Parse()

	valueRange = *max - *min
	nom = valueRange / 2 + *min
	value = randomizer.Float64() * valueRange + *min

	dur, _ := time.ParseDuration(strconv.Itoa(1000/int(*freq)) + "ms")
	signal := time.Tick(dur)
	buf := new(bytes.Buffer)
	sensorName := *name
	qutils.PublishToFanout(qutils.SensorsListExchange, []byte(sensorName))

	go listenToDiscover()

	for range signal {
		reading := dto.SensorMessage{
			Name: *name,
			Value: value,
			Timestamp: time.Now(),
		}
		buf.Reset()
		enc := gob.NewEncoder(buf)
		enc.Encode(reading)
		log.Printf("Reading sent. Value %v\n", value)

		qutils.PublishToDirect("", sensorName, buf.Bytes())

		reCalcValue()
	}
}

func listenToDiscover() {
	q := qutils.DeclareQueue(qutils.SensorsDiscoverExchange, "", true, []string{""})

	qutils.Consume(q.Name, func(msg amqp.Delivery) {
		fmt.Println("Discovery request")
		qutils.PublishToFanout(qutils.SensorsListExchange, []byte(*name))
	})
}

func reCalcValue() {
	var maxStep, minStep float64

	if value < nom {
		maxStep = *stepSize
		minStep = -1 * *stepSize * (value - *min) / (nom - *min)
	} else {
		maxStep = *stepSize * (*max - value) / (*max - nom)
		minStep = -1 * *stepSize
	}

	value += randomizer.Float64() * (maxStep - minStep) + minStep
}
