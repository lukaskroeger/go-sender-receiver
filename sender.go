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

var slidingWindow [slidingWindowSize]message
var slidingWindowIndex int

var ackBuff []byte

var conSend *net.UDPConn
var conRec *net.UDPConn

var timerMap map[int]*time.Timer

func main() {
	arguments := os.Args
	if len(arguments) != 2 {
		fmt.Println("Programm usage: ./server ip:port")
		return
	}

	destination := arguments[1]
	addrSend, err := net.ResolveUDPAddr("udp4", destination)
	checkError(err)

	addrRec, err := net.ResolveUDPAddr("udp4", ":4445")
	checkError(err)

	conSend, err = net.DialUDP("udp4", nil, addrSend)
	checkError(err)
	defer conSend.Close()

	conRec, err = net.ListenUDP("udp4", addrRec)
	checkError(err)
	defer conRec.Close()

	ackBuff = make([]byte, 1024)
	timerMap = make(map[int]*time.Timer)

	slidingWindowIndex = 0
	seqNumber := 0
	until := 0
	for {
		if seqNumber >= slidingWindowSize {
			until = (packageAcknowleged(conRec) + 1) % slidingWindowSize
		}
		for slidingWindowIndex != until || seqNumber < slidingWindowSize {
			message := message{slidingWindowIndex, seqNumber, time.Now()}
			data, _ := json.Marshal(message)
			slidingWindow[slidingWindowIndex] = message
			conSend.Write(data)

			fmt.Println("Transmitt: ", slidingWindowIndex)
			time.Sleep(time.Second * 1)
			slidingWindowIndex = (slidingWindowIndex + 1) % slidingWindowSize
			seqNumber += 1
		}

	}

}

func packageAcknowleged(con *net.UDPConn) int {
	for true {
		data := new(ack)
		n, _, err := con.ReadFromUDP(ackBuff)
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal(ackBuff[:n], data)
		result := data.SlidingWindow
		if data.IsNak {
			retransmittPackage(data.SlidingWindow)
		} else {
			fmt.Println("Ack: ", data.SlidingWindow)
			if timerMap[data.SlidingWindow] != nil {
				timerMap[data.SlidingWindow].Stop()
			}
			return result
		}

	}
	return -1
}

func retransmittPackage(num int) {
	message := slidingWindow[num]
	message.Time = time.Now()
	data, _ := json.Marshal(message)
	conSend.Write(data)
	fmt.Println("Retransmitt: ", num)
	f := func() { retransmittPackage(num) }
	timerMap[num] = time.AfterFunc(time.Second*10, f)
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
