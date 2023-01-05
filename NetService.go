package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// NTPv3 packet is 48 bytes
// NTPv4 packet is 68 bytes

type NTPService struct {
}

func (ntp *NTPService) HandleStanderNTPServer(buf []byte, conn *net.UDPConn, tarAddr *net.UDPAddr) {
	pkt, err := ParseNTPPacket(buf)
	if err != nil {
		fmt.Println("Error parsing packet:", err)
		return
	}

	// Create response packet
	resp, err := CreateNTPResponse(pkt)
	if err != nil {
		fmt.Println("Error creating response packet:", err)
		return
	}
	// Send response
	_, err = conn.WriteToUDP(resp, tarAddr)
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}
}

func (ntp *NTPService) HandleMicroSoftNTPServer(buf []byte, conn *net.UDPConn, tarAddr *net.UDPAddr) {
	pkt, err := ParseMSPacket(buf)
	if err != nil {
		fmt.Println("Error parsing packet:", err)
		return
	}

	// Create response packet
	resp, err := CreateMicroSoftNTPResponse(pkt)
	if err != nil {
		fmt.Println("Error creating response packet:", err)
		return
	}
	// Send response
	_, err = conn.WriteToUDP(resp, tarAddr)
	fmt.Println("发送给%s_____%v完成", tarAddr, resp)
	if err != nil {
		fmt.Println("Error sending response:", err)
		return
	}
}
func ParseMSPacket(buf []byte) (MicrosoftNTPPacket, error) {
	// parse the incoming NTP packet
	var ntp MicrosoftNTPPacket
	ntp.Flags = buf[0]
	ntp.Stratum = buf[1]
	ntp.Poll = buf[2]
	ntp.Precision = buf[3]
	ntp.RootDelay = binary.BigEndian.Uint32(buf[4:8])
	ntp.RootDispersion = binary.BigEndian.Uint32(buf[8:12])
	ntp.ReferenceID = binary.BigEndian.Uint32(buf[12:16])
	ntp.ReferenceTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[16:20])-2208988800), 0)
	ntp.OriginateTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[20:24])-2208988800), 0)
	ntp.ReceiveTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[24:28])-2208988800), 0)
	ntp.TransmitTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[28:32])-2208988800), 0)

	//ntp.ReferenceTimestamp = binary.BigEndian.Uint64(buf[16:24])
	//ntp.OriginateTimestamp = binary.BigEndian.Uint64(buf[24:32])
	//ntp.ReceiveTimestamp = binary.BigEndian.Uint64(buf[32:40])
	//ntp.TransmitTimestamp = binary.BigEndian.Uint64(buf[40:48])
	return ntp, nil
}

// ParseNTPPacket parses an NTP packet
func ParseNTPPacket(buf []byte) (NTPv3Packet, error) {
	/*
		LeapIndicator  byte      //跳跃指示器（LeapIndicator）：2位，指示NTP协议运行的状态，分为正常、提前、延后和未知状态。
		VersionNumber  byte      //版本号（VersionNumber）：3位，表示NTP协议的版本号。
		Mode           byte      //模式（Mode）：3位，指示报文的模式，分为客户端、服务器、广播和多播模式。	1byte
	*/
	var pkt NTPv3Packet
	// Parse packet
	pkt.LeapIndicator = (buf[0] >> 6) & 0x03
	pkt.VersionNumber = (buf[0] >> 3) & 0x07
	pkt.Mode = buf[0] & 0x07
	pkt.Stratum = buf[1]
	pkt.Poll = buf[2]
	pkt.Precision = buf[3]
	pkt.RootDelay = binary.BigEndian.Uint32(buf[4:8])
	pkt.RootDispersion = binary.BigEndian.Uint32(buf[8:12])
	pkt.ReferenceID = binary.BigEndian.Uint32(buf[12:16])
	pkt.ReferenceTime = time.Unix(int64(binary.BigEndian.Uint32(buf[16:20])-2208988800), 0)
	pkt.OriginTime = time.Unix(int64(binary.BigEndian.Uint32(buf[20:24])-2208988800), 0)
	pkt.ReceiveTime = time.Unix(int64(binary.BigEndian.Uint32(buf[24:28])-2208988800), 0)
	pkt.TransmitTime = time.Unix(int64(binary.BigEndian.Uint32(buf[28:32])-2208988800), 0)
	pkt.Authentication = buf[32:40]
	return pkt, nil
}

// CreateNTPResponse creates an NTP response packet
func CreateNTPResponse(pkt NTPv3Packet) ([]byte, error) {
	// Create response packet
	var buf [NtpV4PacketSize]byte
	buf[0] = (pkt.LeapIndicator << 6) | (pkt.VersionNumber << 3) | pkt.Mode
	buf[1] = pkt.Stratum
	buf[2] = pkt.Poll
	buf[3] = pkt.Precision
	binary.BigEndian.PutUint32(buf[4:8], pkt.RootDelay)
	binary.BigEndian.PutUint32(buf[8:12], pkt.RootDispersion)
	binary.BigEndian.PutUint32(buf[12:16], pkt.ReferenceID)
	binary.BigEndian.PutUint32(buf[16:20], uint32(pkt.ReferenceTime.Unix()+2208988800))
	binary.BigEndian.PutUint32(buf[20:24], uint32(pkt.OriginTime.Unix()+2208988800))
	binary.BigEndian.PutUint32(buf[24:28], uint32(pkt.ReceiveTime.Unix()+2208988800))
	// Set transmit time to current time
	transmitTime := time.Now().Unix() + 2208988800
	binary.BigEndian.PutUint32(buf[28:32], uint32(transmitTime))
	buf[32] = 0 // No authentication
	buf[33] = 0
	buf[34] = 0
	buf[35] = 0
	return buf[:], nil
}
func CreateMicroSoftNTPResponse(pkt MicrosoftNTPPacket) ([]byte, error) {
	/*
		// Create response packet
		var buf [NtpV3PacketSize]byte //NTPv3  && MSNTP length =48
		buf[0] = pkt.Flags
		buf[1] = pkt.Stratum
		buf[2] = pkt.Poll
		buf[3] = pkt.Precision
		binary.BigEndian.PutUint32(buf[4:8], pkt.RootDelay)
		binary.BigEndian.PutUint32(buf[8:12], pkt.RootDispersion)
		binary.BigEndian.PutUint32(buf[12:16], pkt.ReferenceID)
		binary.BigEndian.PutUint32(buf[16:24], uint32(pkt.ReferenceTimestamp.Unix()+2208988800))
		binary.BigEndian.PutUint32(buf[24:32], uint32(pkt.OriginateTimestamp.Unix()+2208988800))
		binary.BigEndian.PutUint32(buf[32:40], uint32(pkt.ReceiveTimestamp.Unix()+2208988800))
		// Set transmit time to current time
		transmitTime := time.Now().Unix() + 2208988800
		binary.BigEndian.PutUint32(buf[40:48], uint32(transmitTime))
		buf[32] = 0 // No authentication
		buf[33] = 0
		buf[34] = 0
		buf[35] = 0
		return buf[:], nil
	*/
	// SNTP服务器处理
	// ...

	time.Sleep(time.Second * 1)
	// 获取当前时间
	now := time.Now()

	// 设置SNTP响应头
	resp := make([]byte, 48)
	resp[0] = 0x24 // 响应头
	resp[1] = 0x01

	// 设置Transmit Timestamp
	resp[40] = byte(now.Unix() >> 24)
	resp[41] = byte(now.Unix() >> 16)
	resp[42] = byte(now.Unix() >> 8)
	resp[43] = byte(now.Unix())
	return resp[:], nil
	// 发送响应
	//_, err := conn.WriteToUDP(resp, addr)
	//checkError(err)
}

type MicrosoftNTPPacket struct {
	Flags              byte      //[0]		1byte  client
	Stratum            byte      //[1]		1byte
	Poll               byte      //[2]		1byte
	Precision          byte      //[3]		1byte
	RootDelay          uint32    //[4:8]    4byte
	RootDispersion     uint32    //[8:12]   4byte
	ReferenceID        uint32    //[12:16]  4byte
	ReferenceTimestamp time.Time //uint64//[16:24]  8byte
	OriginateTimestamp time.Time //uint64//[24:32]  8byte
	ReceiveTimestamp   time.Time //uint64//[32:40]  8byte
	TransmitTimestamp  time.Time //uint64//[40:48]  8byte
}

type NTPv3Packet struct {
	LeapIndicator  byte      //跳跃指示器（LeapIndicator）：2位，指示NTP协议运行的状态，分为正常、提前、延后和未知状态。
	VersionNumber  byte      //版本号（VersionNumber）：3位，表示NTP协议的版本号。
	Mode           byte      //模式（Mode）：3位，指示报文的模式，分为客户端、服务器、广播和多播模式。	1byte
	Stratum        byte      //层次（Stratum）：8位，表示报文的传输层次，分为主服务器、从服务器、参考源和其他等。	1byte
	Poll           byte      //轮询时间（Poll）：8位，指示客户端请求服务器发送报文的间隔。	1byte
	Precision      byte      //精度（Precision）：8位，表示报文发送时与UTC时间之间的偏移量的精度。	1byte
	RootDelay      uint32    //根延迟（RootDelay）：32位，表示客户端与服务器之间的网络延迟。	4byte
	RootDispersion uint32    //根失望（RootDispersion）：32位，表示客户端与服务器的时间差。	4byte
	ReferenceID    uint32    //参考ID（ReferenceID）：32位，表示参考时间源的ID。	4byte
	ReferenceTime  time.Time //参考时间（ReferenceTime）：参考时间，表示参考时间源的时间。 8byte
	OriginTime     time.Time //原始时间（OriginTime）：原始时间，表示报文发出时的时间。	8byte
	ReceiveTime    time.Time //接收时间（ReceiveTime）：接收时间，表示报文接收时的时间。	8byte
	TransmitTime   time.Time //发送时间（TransmitTime）：发送时间，表示报文发送时的时间。	8byte
	Authentication []byte    //认证（Authentication）：32位，用于对报文进行认证和安全传输。	4byte
}

// NTPv4 structure
type NTPv4Packet struct {
	LeapIndicator     byte     //跳跃指示器（LeapIndicator）：2bit，指示NTP协议运行的状态，分为正常、提前、延后和未知状态。
	Version           uint8    //NTP版本号：3bit，用来指示使用的NTP版本号
	Mode              uint8    //模式：3bit，用来指示NTP客户端和服务器之间的交互方式
	Stratum           uint8    //NTP服务器的级别：8bit，用来指示NTP服务器的级别
	PollInterval      uint8    //客户端向服务器查询时间的间隔：8bit，用来指示客户端向服务器发送请求的间隔
	Precision         int8     //NTP服务器时间的精度：8bit，用来指示NTP服务器时间的精度
	RootDelay         int32    //根延迟：32bit，用来指示NTP客户端和服务器之间的延迟
	RootDisp          int32    //根分散：32bit，用来指示NTP客户端和服务器之间的时间分散
	ReferenceID       uint32   //参考标识符：32bit，用来指示参考时钟的标识符
	RefTimestamp      uint64   //参考时间戳：64bit，用来指示服务器的参考时间
	OrigTimestamp     uint64   //发出时间戳：64bit，用来指示客户端发出请求的时间
	RecvTimestamp     uint64   //接收时间戳：64bit，用来指示服务器接收请求的时间
	TransmitTimestamp uint64   //发送时间戳：64bit，用来指示服务器发送响应的时间
	Auth              [8]uint8 //Authentication字段：64bit，用来验证报文的可靠性
}

// 判断是否为标准NTP请求报文
func IsStandardNtpRequest(pkt []byte) bool {

	return (pkt[1] == 3) //pkt[0]=219 二进制为11011011
}

// 判断是否为Microsoft NTP请求报文
func IsMicrosoftNtpRequest(pkt []byte) bool {
	return (pkt[2] == 7)
}
