package main

import (
	"fmt"
	"net"
)

const (
	NtpV4PacketSize = 48 //MS-NTP  && NTP-v3 v4
	NtpV3PacketSize = 68
)

func main() {
	// Create a UDP connection
	var netservice = NTPService{}
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		//IP:   net.IPv4zero,
		IP:   net.ParseIP("192.168.16.120"),
		Port: 123,
	})
	if err != nil {
		panic(err)
	}

	// Close connection
	defer conn.Close()

	fmt.Println("Listening for NTP packets...")

	// Buffer for incoming data
	buf := make([]byte, NtpV4PacketSize)

	// Wait for incoming packets
	for {
		fmt.Println("for内部循环监听中")
		// Read data from socket
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		fmt.Println(addr)
		// Check packet size
		if n != NtpV3PacketSize && n != NtpV4PacketSize {
			fmt.Println("Invalid packet size")
			continue
		}
		//长度判断完毕进行版本识别
		// Parse packet--recoginze standNTP or MSNTP
		if IsStandardNtpRequest(buf[0:n]) {
			//先尝试标准NTP不行MSNTP 再不行抛弃
			//如果是通用标准NTP服务调用解析服务
			netservice.HandleStanderNTPServer(buf, conn, addr)
			continue
		} else {
			//if IsMicrosoftNtpRequest(buf[0:n]) {
			netservice.HandleMicroSoftNTPServer(buf, conn, addr)
			continue
		}

	}
}
