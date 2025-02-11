package main

import (
	"fmt"
	"net"
	"sort"
)

func worker(ports, results chan int) {
	for p := range ports {
		address := fmt.Sprintf("127.0.0.1:%d", p)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			results <- 0
			continue
		}
		conn.Close()
		results <- p
	}
}

func main() {
	portsToScan := 65535
	ports := make(chan int, 100)
	results := make(chan int)
	var openports []int

	for i := 1; i <= cap(ports); i++ {
		go worker(ports, results)
	}

	go func() {
		for i := 1; i <= portsToScan; i++ {
			ports <- i
		}
	}()

	for i := 0; i < portsToScan; i++ {
		port := <-results
		if port != 0 {
			openports = append(openports, port)
		}
	}

	close(ports)
	close(results)
	sort.Ints(openports)
	for _, port := range openports {
		fmt.Printf("%d is open!\n", port)
	}
}
