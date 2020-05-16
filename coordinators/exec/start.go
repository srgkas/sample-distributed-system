package main

import (
	"first-distributed-system/coordinators"
	"fmt"
)

func main()  {
	fmt.Println("Starting coordinator")
	l := coordinators.NewSensorsListener()
	go l.ListenForSensors()
	go l.ListenForWebAppDiscovery()
	l.DiscoverSensors()

	var a string
	fmt.Scan(&a)
}
