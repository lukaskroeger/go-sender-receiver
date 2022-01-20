package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/bits-and-blooms/bloom"
)

type message struct {
	//header Data
	SlidingWindow int
	ListIndex     int
	Time          time.Time
	//payload data
	SeqNum  int
	Payload []byte
}

type ack struct {
	SlidingWindow int
	Bloom         []byte
}

const slidingWindowSize = 1000

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
	err = conRec.SetReadBuffer(100000 * 1024)
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
		//fmt.Println("waiting")
		n, conSendAddr, _ := conRec.ReadFromUDP(buffer)
		//fmt.Println(string(buffer), n)
		transferedTime += time.Since(lastPackTime)
		transferedData += 8 * n
		transfaredPackages += 1
		if mode == "complete" {
			json.Unmarshal(buffer[:n], data)
			diff := data.ListIndex - lastPackNum
			if diff > 1 {
				lostPackages += diff - 1
				for n := lastPackNum + 1; n < lastPackNum+diff; n++ {
					missingPackages = append(missingPackages, n)
				}
			}

			if data.ListIndex+1 == slidingWindowSize {
				//fmt.Println(missingPackages)
				sendNacks(conSendAddr, data.SlidingWindow)
			}
			latencySum += time.Since(data.Time).Seconds()
			lastPackNum = data.ListIndex
		}
		output(data)
		lastPackTime = time.Now()
	}

}

func sendNacks(addr *net.UDPAddr, slidingWindow int) {
	//fmt.Println("nack send")
	bloomFilter := bloom.New(slidingWindowSize, 5)
	for _, pack := range missingPackages {
		bloomFilter.Add([]byte(fmt.Sprint(pack)))
	}
	addrSend, err := net.ResolveUDPAddr("udp4", addr.IP.String()+":4445")
	checkError(err)

	con, err := net.DialUDP("udp4", nil, addrSend)
	checkError(err)
	defer con.Close()
	bloom, _ := bloomFilter.MarshalJSON()
	ackPackage := ack{SlidingWindow: slidingWindow, Bloom: bloom}

	data, _ := json.Marshal(ackPackage)
	con.Write(data)
	missingPackages = missingPackages[:0]
	//fmt.Println("nack sent")
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
