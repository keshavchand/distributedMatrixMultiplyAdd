package main

import (
	"bufio"
	"log"
	"net"
	"time"
)

// TODO:
// load balancing
// scheduling
// increasing throughput

var workers []net.Conn
func main() {
	registerWorkers()
}

func registerWorkers() {
	serverUrl := ":7080"
	conn, err := net.Listen("tcp", serverUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for {
		client, err := conn.Accept()
		if err != nil {
			log.Println(err)
			break
		}

		go func(client net.Conn) {
			defer client.Close()
			monitorHeartbest(client)
		}(client)
	}
}

func monitorHeartbest(client net.Conn) {
	scanner := bufio.NewScanner(client)
	if !scanner.Scan() {
		log.Println("No Reconnect Address")
		return
	}

	url := scanner.Text()
	conn, err := net.Dial("tcp", url)
	if err != nil {
		log.Println("connection to client", err)
		return
	}
	defer conn.Close()

	beat := make(chan time.Time, 1)
	go func() {
		defer close(beat)
		for scanner.Scan() {
			beat <- time.Now()
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	lastBeat := time.Now()
	for {
		select {
		case t := <-ticker.C:
			if t.Sub(lastBeat) >= 10*time.Second {
				log.Println("Suspecting Worker Died: ", url)
				return
			}
		case t, open := <-beat:
			if !open{
				break
			}
			log.Println("Beat At: ", t)
			lastBeat = t
		}
	}
}
