package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

type message struct {
	SlidingWindow int
	SeqNum        int
	Time          time.Time
}
type ack struct {
	SlidingWindow int
	IsNak         bool
}

const slidingWindowSize = 10000

var lastTime time.Time
var transferedData int
var transferedTime time.Duration
var latencySum float64
var transfaredPackages int
var mode string
var lostPackages int

var missingPackages []int

var tempRetransmitts int

func main() {
	arguments := os.Args
	if len(arguments) != 3 {
		fmt.Println("Programm usage: ./server <mode> port")
		fmt.Println("mode: simple/complete")
	}

	port := ":" + arguments[2]
	mode = arguments[1]
	fmt.Printf("Mode: %s \n", mode)

	addr, err := net.ResolveUDPAddr("udp4", port)
	checkError(err)

	conRec, err := net.ListenUDP("udp4", addr)
	checkError(err)
	defer conRec.Close()

	lastTime = time.Now()
	transferedData = 0
	latencySum = float64(0)
	transfaredPackages = 0
	lostPackages = 0

	buffer := make([]byte, 1024)
	lastPackTime := time.Now()
	lastPackNum := 0
	data := new(message)
	missingPackages = make([]int, 0)
	tempRetransmitts = 0

	for {
		n, conSendAddr, _ := conRec.ReadFromUDP(buffer)
		transferedTime += time.Since(lastPackTime)
		transferedData += 8 * n
		transfaredPackages += 1
		if mode == "complete" {
			json.Unmarshal(buffer[:n], data)
			if removeFromOrderedIfThere(data.SlidingWindow, missingPackages) {
				sendAcknowledgement(conSendAddr, data.SlidingWindow)
				tempRetransmitts += 1
			} else {
				diff := (data.SlidingWindow - lastPackNum) % slidingWindowSize
				if diff > 1 {
					lostPackages += diff
					sendNotAck(conSendAddr, lastPackNum+1, diff)
					for n := lastPackNum + 1; n < lastPackNum+diff; n++ {
						missingPackages = append(missingPackages, n%slidingWindowSize)
					}
				} else {
					sendAcknowledgement(conSendAddr, data.SlidingWindow)
				}
			}
			latencySum += time.Since(data.Time).Seconds()
			lastPackNum = data.SlidingWindow

		}
		output(data)
		lastPackTime = time.Now()
	}

}

func sendAcknowledgement(addr *net.UDPAddr, seqNumber int) {
	addrSend, err := net.ResolveUDPAddr("udp4", addr.IP.String()+":4445")
	checkError(err)

	con, err := net.DialUDP("udp4", nil, addrSend)
	checkError(err)
	defer con.Close()

	ackPackage := ack{IsNak: false}
	if len(missingPackages) > 0 {
		ackPackage.SlidingWindow = (missingPackages[0] - 1) % slidingWindowSize
	} else {
		ackPackage.SlidingWindow = seqNumber
	}
	data, _ := json.Marshal(ackPackage)
	con.Write(data)
}

func sendNotAck(addr *net.UDPAddr, from int, count int) {
	addrSend, err := net.ResolveUDPAddr("udp4", addr.IP.String()+":4445")
	checkError(err)

	con, err := net.DialUDP("udp4", nil, addrSend)
	checkError(err)
	defer con.Close()
	for num := from; num <= from+count; num++ {
		data, err := json.Marshal(ack{num % slidingWindowSize, true})
		checkError(err)
		con.Write(data)
	}
}

func output(data *message) {
	if time.Since(lastTime).Seconds() > 1 {
		if mode == "complete" {
			complexOutput(transfaredPackages, transferedData, transferedTime, data.SeqNum, latencySum, lostPackages)
		} else {
			simpleOutput(transfaredPackages, transferedData, transferedTime)
		}
		lastTime = time.Now()
		transferedData = 0
		transferedTime = time.Duration(0)
		latencySum = float64(0)
		transfaredPackages = 0
	}
}

func simpleOutput(transferedPackages int, transferedData int, transferedTime time.Duration) {
	fmt.Printf("Package count: %d \n", transferedPackages)
	fmt.Printf("Datarate : %f Mbit/s\n", float64(transferedData)/transferedTime.Seconds()/1024/1024)
	fmt.Print("\033[A")
	fmt.Print("\033[A")
}

func complexOutput(transferedPackages int, transferedData int, transferedTime time.Duration, seqNumber int, latencySum float64, lostPackages int) {
	fmt.Printf("Package count: %d /s\n", transferedPackages)
	fmt.Printf("Number: %d \n", seqNumber)
	fmt.Printf("Datarate : %f Mbit/s\n", float64(transferedData)/transferedTime.Seconds()/1024/1024)
	fmt.Printf("Latency: %f \n", latencySum/float64(transferedPackages))
	fmt.Printf("Lost packages: %d \n", lostPackages)
	fmt.Printf("Currenly missing: %d \n", len(missingPackages))
	fmt.Printf("Currenly missing: %d \n", tempRetransmitts)
	fmt.Print("\033[A\033[A\033[A\033[A\033[A\033[A\033[A")
}

func removeFromOrderedIfThere(value int, list []int) bool {
	for i := 0; i < len(list); i++ {
		if list[i] == value {
			list = append(list[:i], list[i+1:]...)
			return true
		}
		if list[i] > value {
			return false
		}

	}
	return false
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
