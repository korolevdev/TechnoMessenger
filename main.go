package main

import (
	"./server"
)

// PORT of Server
const PORT = 7788

// Main function
func main() {
	s := server.CreateInstance()
	s.Start(PORT)
}
