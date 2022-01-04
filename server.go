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
        fmt.Println("Programm usage: ./server port")
    }

    port := ":"+arguments[1]

    addr, err := net.ResolveUDPAddr("udp4", port)
    checkError(err)

    con, err := net.ListenUDP("udp4", addr)
    checkError(err)

    defer con.Close()

    buffer := make([]byte, 1024)
    lastPackTime := time.Now()
    lastPackNum := 0
    lostPackages := 0
    for {
        n, _, _ := con.ReadFromUDP(buffer)
        data := new(message)
        json.Unmarshal(buffer[:n], data)
        fmt.Printf("Number: %d \n", data.SeqNum)
        fmt.Printf("Latency: %s \n", time.Since(data.Time))
        seconds := time.Since(lastPackTime).Seconds()
        fmt.Printf("Datarate : %f Mbit/s\n" , float64(8*n)/seconds/1024/1024)

        if (data.SeqNum - lastPackNum) > 1 {
            lostPackages += data.SeqNum - lastPackNum - 1
        }
        lastPackNum = data.SeqNum
        fmt.Printf("Lost packages: %d \n", lostPackages)
        lastPackTime = time.Now()
        fmt.Print("\033[A")
        fmt.Print("\033[A")
        fmt.Print("\033[A")
        fmt.Print("\033[A")
    }

}

func checkError(err error){
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
