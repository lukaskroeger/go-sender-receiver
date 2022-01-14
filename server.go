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
	if len(arguments) != 3 {
		fmt.Println("Programm usage: ./server <mode> port")
		fmt.Println("mode: simple/complete")
	}

	port := ":" + arguments[2]
	mode := arguments[1]
	fmt.Printf("Mode: %s \n", mode)

	addr, err := net.ResolveUDPAddr("udp4", port)
	checkError(err)

	con, err := net.ListenUDP("udp4", addr)
	checkError(err)

	defer con.Close()

	buffer := make([]byte, 3000)
	lastPackTime := time.Now()
	lastPackNum := 0
	lostPackages := 0
	lastTime := time.Now()
	transferedData := 0
	var transferedTime time.Duration
	latencySum := float64(0)
	transfaredPackages := 0
	var data message
	for {
		n, _, _ := con.ReadFromUDP(buffer)
		transferedTime += time.Since(lastPackTime)
		transferedData += 8 * n
		transfaredPackages += 1
		if mode == "complete" {
			data := new(message)
			json.Unmarshal(buffer[:n], data)
			if (data.SeqNum - lastPackNum) > 1 {
				lostPackages += data.SeqNum - lastPackNum - 1
			}
			latencySum += time.Since(data.Time).Seconds()
		}
		if time.Since(lastTime).Seconds() > 1 {

			if mode == "complete" {
				complexOutput(transfaredPackages, transferedData, transferedTime, data, latencySum, lostPackages)
			} else {
				simpleOutput(transfaredPackages, transferedData, transferedTime)
			}
			lastTime = time.Now()
			transferedData = 0
			transferedTime = time.Duration(0)
			latencySum = float64(0)
			transfaredPackages = 0
		}
		lastPackTime = time.Now()
	}

}

func simpleOutput(transferedPackages int, transferedData int, transferedTime time.Duration) {
	fmt.Printf("Package count: %d \n", transferedPackages)
	fmt.Printf("Datarate : %f Mbit/s\n", float64(transferedData)/transferedTime.Seconds()/1024/1024)
	fmt.Print("\033[A")
	fmt.Print("\033[A")
}

func complexOutput(transferedPackages int, transferedData int, transferedTime time.Duration, data message, latencySum float64, lostPackages int) {
	fmt.Printf("Package count: %d /s\n", transferedPackages)
	fmt.Printf("Number: %d \n", data.SeqNum)
	fmt.Printf("Datarate : %f Mbit/s\n", float64(transferedData)/transferedTime.Seconds()/1024/1024)
	fmt.Printf("Latency: %f \n", latencySum/float64(transferedPackages))
	fmt.Printf("Lost packages: %d \n", lostPackages)
	fmt.Print("\033[A\033[A\033[A\033[A\033[A")
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
