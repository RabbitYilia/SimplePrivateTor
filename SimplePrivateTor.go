package main

import (
	"bufio"
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var peermap []string

func main() {
	peermap = append(peermap, "127.0.0.1:6161")
	ListenAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 6161,
	}
	conn, err := net.ListenUDP("udp", ListenAddr)
	defer conn.Close()
	if err != nil {
		log.Fatal(err)
	}

	go ProcessRX(conn)

	for {
		log.Println("Please input dst:")
		inputReader := bufio.NewReader(os.Stdin)
		input, err := inputReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.Trim(input, "\n")
		input = strings.Trim(input, "\r")
		if input == "" {
			break
		}
		TXData := make(map[string]string)
		TXData["DST"] = input
		TXData["Timestamp"] = strconv.FormatInt(time.Now().UnixNano(), 10)
		TXData["TTL"] = strconv.Itoa(RandInt(1, 10))
		log.Println("Please input msg:")
		inputReader = bufio.NewReader(os.Stdin)
		input, err = inputReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.Trim(input, "\n")
		input = strings.Trim(input, "\r")
		if input == "" {
			break
		}
		TXData["Data"] = input
		HandletoPeer(TXData)
	}
}

func ProcessRX(conn *net.UDPConn) {
	RXByte := make([]byte, 4096)
	RXData := make(map[string]string)
	for {
		read, _, err := conn.ReadFromUDP(RXByte)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(RXByte[:read], &RXData)
		if err != nil {
			return
		}
		TTL, err := strconv.Atoi(RXData["TTL"])
		if err != nil {
			return
		}
		if TTL == 0 {
			log.Println(RXData["Data"])
		}
		if TTL == 1 {
			HandletoDST(RXData)
		} else {
			HandletoPeer(RXData)
		}
	}
}

func HandletoPeer(Data map[string]string) {
	TTL, err := strconv.Atoi(Data["TTL"])
	if err != nil {
		return
	}
	var DstAddr []string
	if strings.Index(Data["DST"], ":") != strings.LastIndex(Data["DST"], ":") {
		DstAddr = strings.Split(strings.Replace(Data["DST"], "[", "", 0), "]:")
	} else {
		DstAddr = strings.Split(Data["DST"], ":")
	}
	DSTPort, err := strconv.Atoi(DstAddr[1])
	if err != nil {
		return
	}
	DST := &net.UDPAddr{
		IP:   net.ParseIP(DstAddr[0]),
		Port: DSTPort,
	}
	Data["TTL"] = strconv.Itoa(TTL - 1)
	SendJson, err := json.Marshal(Data)
	if err != nil {
		log.Fatal(err)
	}
	socket, err := net.DialUDP("udp", nil, DST)
	if err != nil {
		return
	}
	socket.Write([]byte(SendJson))
	socket.Close()
}

func HandletoDST(Data map[string]string) {
	DSTnum := RandInt(0, len(peermap)-1)
	var DstAddr []string
	if strings.Index(peermap[DSTnum], ":") != strings.LastIndex(peermap[DSTnum], ":") {
		DstAddr = strings.Split(strings.Replace(peermap[DSTnum], "[", "", 0), "]:")
	} else {
		DstAddr = strings.Split(peermap[DSTnum], ":")
	}
	DSTPort, err := strconv.Atoi(DstAddr[1])
	if err != nil {
		return
	}
	DST := &net.UDPAddr{
		IP:   net.ParseIP(DstAddr[0]),
		Port: DSTPort,
	}
	Data["DST"] = "0"
	SendJson, err := json.Marshal(Data)
	if err != nil {
		log.Fatal(err)
	}
	socket, err := net.DialUDP("udp", nil, DST)
	if err != nil {
		return
	}
	socket.Write([]byte(SendJson))
	socket.Close()
}

func RandInt(min, max int) int {
	rand.Seed(time.Now().UnixNano() * rand.Int63n(100))
	return min + rand.Intn(max-min+1)
}
