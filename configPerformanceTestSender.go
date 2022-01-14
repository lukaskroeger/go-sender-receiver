package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type message struct {
	SeqNum int
	Time   time.Time
	Data   []byte
}

func main() {
	arguments := os.Args
	if len(arguments) != 5 {
		fmt.Println("Programm usage: ./server ip:port <useJson> <messageSize> <useBufIo>")
		fmt.Println("useJson: true/false => specifies if json encoding should take place in every loop cycle")
		fmt.Println("messageSize: bytes => Size of message data")
		fmt.Println("useBufIo: true/false => use a bufferd IO writer or not")
		return
	}

	destination := arguments[1]
	addr, err := net.ResolveUDPAddr("udp4", destination)
	checkError(err)
	con, err := net.DialUDP("udp4", nil, addr)
	checkError(err)
	defer con.Close()

	var writer io.Writer
	if arguments[4] == "true" {
		writer = bufio.NewWriter(con)
	} else {
		writer = con
	}

	var messageDataSize int
	_, err = fmt.Sscan(arguments[3], &messageDataSize)
	checkError(err)

	messageData := make([]byte, messageDataSize)
	rand.Read(messageData)

	if arguments[2] == "true" {
		sendWithJson(messageData, writer)
	} else {
		sendWithoutJson(messageData, writer)
	}

}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func sendWithJson(messageData []byte, writer io.Writer) {
	for {
		data, _ := json.Marshal(message{0, time.Now(), messageData})
		writer.Write(data)
	}
}

func sendWithoutJson(messageData []byte, writer io.Writer) {
	data, _ := json.Marshal(message{0, time.Now(), messageData})
	for {
		writer.Write(data)
	}
}
