package main

import (
	"crypto/rand"
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

var slidingWindows [3][slidingWindowSize]message
var slidingWindowIndex int

var ackBuff []byte

var conSend *net.UDPConn
var conRec *net.UDPConn

var timerMap map[int]*time.Timer

var palyoadData []byte

func main() {
	arguments := os.Args
	if len(arguments) != 2 {
		fmt.Println("Programm usage: ./server ip:port")
		return
	}
	test := bloom.BloomFilter{}
	fmt.Println(test)
	destination := arguments[1]
	addrSend, err := net.ResolveUDPAddr("udp4", destination)
	checkError(err)

	addrRec, err := net.ResolveUDPAddr("udp4", ":4445")
	checkError(err)

	conSend, err = net.DialUDP("udp4", nil, addrSend)
	checkError(err)
	err = conSend.SetWriteBuffer(100000 * 1024)
	checkError(err)
	defer conSend.Close()

	conRec, err = net.ListenUDP("udp4", addrRec)
	checkError(err)
	defer conRec.Close()

	ackBuff = make([]byte, 1024)
	timerMap = make(map[int]*time.Timer)

	palyoadData := make([]byte, 200)
	rand.Read(palyoadData)

	slidingWindowIndex = 0
	seqNumber := 0
	//until := 0
	var retransmittPackages []message
	for {
		for i, pack := range retransmittPackages {
			//fmt.Println("Restransmitt ", len(retransmittPackages))
			pack.Time = time.Now()
			pack.ListIndex = i
			pack.SlidingWindow = slidingWindowIndex
			data, _ := json.Marshal(pack)
			slidingWindows[slidingWindowIndex][i] = pack
			conSend.Write(data)
			time.Sleep(time.Nanosecond * 500)
		}
		fmt.Println("here", len(retransmittPackages))
		for i := len(retransmittPackages); i < slidingWindowSize; i++ {
			message := message{slidingWindowIndex, i, time.Now(), seqNumber, palyoadData}
			data, _ := json.Marshal(message)
			slidingWindows[slidingWindowIndex][i] = message
			conSend.Write(data)
			seqNumber += 1
			fmt.Println(seqNumber)
			time.Sleep(time.Nanosecond * 500)
		}
		fmt.Println(slidingWindowIndex)
		slidingWindowIndex = (slidingWindowIndex + 1) % 3
		if seqNumber > 2*slidingWindowSize {
			//retransmittPackages = packageAcknowleged(conRec)
		} else {
			time.Sleep(time.Millisecond * 5)
		}
	}

}

func packageAcknowleged(con *net.UDPConn) []message {
	data := new(ack)
	n, _, err := con.ReadFromUDP(ackBuff)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(ackBuff))
	json.Unmarshal(ackBuff[:n], data)
	var toRetransmitt []message
	bloomFilter := bloom.New(slidingWindowSize, 5)
	bloomFilter.UnmarshalJSON(data.Bloom)
	for i := 0; i < slidingWindowSize; i++ {
		if bloomFilter.Test([]byte(fmt.Sprint(slidingWindows[data.SlidingWindow][i].ListIndex))) {
			toRetransmitt = append(toRetransmitt, slidingWindows[data.SlidingWindow][i])
		}
	}
	return toRetransmitt
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
