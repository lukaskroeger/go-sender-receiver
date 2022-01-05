package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

type message struct {
	SeqNum int
	Time   time.Time
}

func main() {
	arguments := os.Args
	if len(arguments) != 2 {
		fmt.Println("Programm usage: ./server port")
	}

	port := ":" + arguments[1]

	addr, err := net.ResolveUDPAddr("udp4", port)
	checkError(err)

	con, err := net.ListenUDP("udp4", addr)
	checkError(err)

	defer con.Close()

	buffer := make([]byte, 1024)
	lastPackTime := time.Now()
	lastPackNum := 0
	lostPackages := 0
	lastTime := time.Now()
	transferedData := 0
	var transferedTime time.Duration
	latencySum := float64(0)
	transfaredPackages := 0
	for {
		n, _, _ := con.ReadFromUDP(buffer)
		data := new(message)
		json.Unmarshal(buffer[:n], data)

		transferedTime += time.Since(lastPackTime)
		transferedData += 8 * n
		if (data.SeqNum - lastPackNum) > 1 {
			lostPackages += data.SeqNum - lastPackNum - 1
		}

		latencySum += time.Since(data.Time).Seconds()
		transfaredPackages += 1

		if time.Since(lastTime).Seconds() > 1 {
			fmt.Printf("Number: %d \n", data.SeqNum)
			fmt.Printf("Datarate : %f Mbit/s\n", float64(transferedData)/transferedTime.Seconds()/1024/1024)
			fmt.Printf("Latency: %f \n", latencySum/float64(transfaredPackages))
			fmt.Printf("Lost packages: %d \n", lostPackages)
			fmt.Print("\033[A")
			fmt.Print("\033[A")
			fmt.Print("\033[A")
			fmt.Print("\033[A")
			lastTime = time.Now()
			transferedData = 0
			transferedTime = time.Duration(0)
			latencySum = float64(0)
			transfaredPackages = 0
		}
		lastPackTime = time.Now()
		lastPackNum = data.SeqNum
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
