package main

import (
	"first-distributed-system/datamanager"
	"fmt"
)

func main() {
	sr := datamanager.NewSensorReader()
	go sr.ReadSensors()

	var a string
	fmt.Scan(&a)
}
