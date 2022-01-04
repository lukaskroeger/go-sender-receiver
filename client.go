package main

import (
    "fmt"
    "net"
    "os"
    "encoding/json"
    "time"
)

type message struct{
    SeqNum int
    Time time.Time
}

func main(){
    arguments := os.Args
    if len(arguments) != 2{
        fmt.Println("Programm usage: ./server ip:port")
        return
    }

    destination := arguments[1]
    addr, err := net.ResolveUDPAddr("udp4", destination)
    checkError(err)

    con, err := net.DialUDP("udp4", nil, addr)
    checkError(err)

    defer con.Close()
    seqNumber := 0
    for{
        data, _ := json.Marshal(message{seqNumber, time.Now()})
        con.Write(data)
        seqNumber += 1
    }

}

func checkError(err error){
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
