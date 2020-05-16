package datamanager

import (
	"bytes"
	"encoding/gob"
	"first-distributed-system/dto"
	"first-distributed-system/qutils"
	"github.com/streadway/amqp"
	"log"
)

type SensorReader struct {
	sensors map[string]int
}

func NewSensorReader() SensorReader {
	return SensorReader{
		sensors: make(map[string]int),
	}
}

type sensor struct {
	ID   int
	Name string
}

func (sr *SensorReader) ReadSensors() {
	qutils.Consume(qutils.QueuePersistentValues, func(msg amqp.Delivery) {
		buf := bytes.NewReader(msg.Body)
		dec := gob.NewDecoder(buf)
		data := dto.SensorMessage{}
		dec.Decode(&data)

		sr.saveData(data)
	})
}

func (sr *SensorReader) saveData(s dto.SensorMessage) {
	if sr.sensors[s.Name] == 0 {
		sr.fetchSensors()
	}

	if sr.sensors[s.Name] == 0 {
		sr.createSensor(sensor{Name: s.Name})
		sr.fetchSensors()
	}

	err := sr.createSensorRecord(sr.sensors[s.Name], s)

	if err != nil {
		panic(err.Error())
	} else {
		log.Printf("Record saved for %s\n", s.Name)
	}
}

func (sr *SensorReader) fetchSensors() error {
	rows, err := connection.Query("SELECT id, name from sensors")

	if err != nil {
		return err
	}

	sr.sensors = make(map[string]int)

	for rows.Next() {
		var name string
		var id int
		rows.Scan(&id, &name)

		sr.sensors[name] = id
	}

	return nil
}

func (sr *SensorReader) createSensor(data sensor) error {
	sql := "INSERT INTO sensors (name) VALUES ($1)"
	_, err := connection.Exec(sql, data.Name)

	if err != nil {
		return err
	}

	return nil
}

func (sr *SensorReader) createSensorRecord(sensorId int, s dto.SensorMessage) error {
	sql := "INSERT INTO sensor_values (sensor_id, value, read_at ) VALUES ($1, $2, $3)"

	_, err := connection.Exec(sql, sensorId, s.Value, s.Timestamp)

	if err != nil {
		return err
	}

	return nil
}
