package main

import (
	"first-distributed-system/web/controller"
	"net/http"
)

func main() {
	controller.Initialize()

	http.ListenAndServe(":3000", nil)
}
