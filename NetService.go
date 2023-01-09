package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

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

// ParseNTPPacket parses an NTP packet
func ParseNTPPacket(buf []byte) (NTPv4Packet, error) {
	/*
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
	*/
	var pkt NTPv4Packet
	// Parse packet
	//假设初始值位0x23 二进制位（0b 00100011） wireshark抓包后识别 0-1 LI 位00 2-4位LN 100 5-7位Mode 11
	//LeapIndicator 跳跃指示器（LeapIndicator）：2bit，指示NTP协议运行的状态，分为正常、提前、延后和未知状态。
	pkt.LeapIndicator = (buf[0] >> 6) //向右移动6位左侧剩余2位 就是头2位  |00000011=3
	//VersionNumber NTP版本号：3bit，用来指示使用的NTP版本号
	var temp = (buf[0] >> 4) //0b1101
	pkt.Version = (temp & (0b111)) >> 1
	pkt.Mode = buf[0] & (0b111)
	pkt.Stratum = buf[1]                                                                               //NTP服务器的级别：8bit，用来指示NTP服务器的级别
	pkt.PollInterval = buf[2]                                                                          //客户端向服务器查询时间的间隔：8bit，用来指示客户端向服务器发送请求的间隔
	pkt.Precision = buf[3]                                                                             ////NTP服务器时间的精度：8bit，用来指示NTP服务器时间的精度 1.0seconed=0b000000
	pkt.RootDelay = binary.BigEndian.Uint32(buf[4:8])                                                  //根延迟：4byte=32bit，用来指示NTP客户端和服务器之间的延迟
	pkt.RootDisp = binary.BigEndian.Uint32(buf[8:12])                                                  //根分散：32bit，用来指示NTP客户端和服务器之间的时间分散
	pkt.ReferenceID = binary.BigEndian.Uint32(buf[12:16])                                              //参考标识符：32bit，用来指示参考时钟的标识符https://www.rfc-editor.org/rfc/rfc5905 搜索关键字"Reference ID" 3/17
	pkt.RefTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[16:24])-2208988800), 0).Unix()      //参考时间戳：64bit，用来指示服务器的参考时间
	pkt.OrigTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[24:32])-2208988800), 0).Unix()     //发出时间戳：64bit，用来指示客户端发出请求的时间
	pkt.RecvTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[32:40])-2208988800), 0).Unix()     //接收时间戳：64bit，用来指示服务器接收请求的时间
	pkt.TransmitTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[40:48])-2208988800), 0).Unix() //发送时间戳：64bit，用来指示服务器发送响应的时间
	return pkt, nil
}

func CreateNTPResponse(pkt NTPv4Packet) ([]byte, error) {
	// Create response packet
	var buf [NtpV4PacketSize]byte
	/*
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
	*/
	//buf[0]最后3位 客户端3 服务端4  client 00 100 011 server 00 100

	buf[0] = buf[0] & 0b11111100
	var test byte = 0b11111100
	fmt.Println(test << 6 >> 6)
	//buf[0] = (pkt.LeapIndicator << 6) | (pkt.Version << 3) | pkt.Mode
	buf[1] = 0b00000011 //pkt.Stratum //服务器时间级别自定义为3 0b00000011=3  来时其他 2 1 级同步结果
	buf[2] = 0b00000000 //pkt.PollInterval 可以设置为0对服务端而言 该值无意义
	buf[3] = 0b00000000 //pkt.Precision NTP服务器时间的精度：8bit，用来指示NTP服务器时间的精度
	copy(buf[4:8], []byte{0, 0, 0, 0})
	//binary.BigEndian.PutUint32(buf[4:8], uint32(pkt.RootDelay)) //根延迟：32bit，用来指示NTP客户端和服务器之间的延迟 RootDelay值 不包含网络传输所需时间，而是由NTP服务器从四个精确的NTP服务器获取标准时间，并计算出与服务器的延迟，然后根据计算出的延迟值计算出RootDelay的值。---RootDelay值取决于NTP服务器之间的网络延迟，如果没有其他NTP服务器可参考，可以将RootDelay值设置为0或较小的值。
	copy(buf[8:12], []byte{0, 0, 0, 0})
	//binary.BigEndian.PutUint32(buf[8:12], uint32(pkt.RootDisp)) //RootDisp值是根据RootDelay值计算出来的，如果RootDelay值设置为0或较小的值，那么RootDisp值也会很小。
	//localserverip:=net.Conn.LocalAddr()
	copy(buf[12:16], net.IPv4(192, 168, 16, 120))
	//binary.BigEndian.PutUint32(buf[12:16], pkt.ReferenceID)//一般为服务器ipv4
	//binary.BigEndian.PutUint32(buf[16:20], uint32(pkt.RefTimestamp+2208988800))// RefTimestamp可以参考通用的NTP服务器中的时间戳，如果没有可用的NTP服务器，可以使用服务端系统的当前时间戳。
	copy(buf[16:24], time.Now().String())
	binary.BigEndian.PutUint32(buf[24:32], uint32(pkt.OrigTimestamp+2208988800)) //客户端时间戳 OrigTimestamp+2208988800的意思是将OrigTimestamp的时间戳从1970年1月1日0点转换为1900年1月1日0点的时间戳，NTP协议使用的是1900年1月1日0点的时间戳。
	binary.BigEndian.PutUint32(buf[32:40], uint32(pkt.RecvTimestamp+2208988800))
	// Set transmit time to current time
	transmitTime := time.Now().Unix() + 2208988800
	binary.BigEndian.PutUint32(buf[40:48], uint32(transmitTime))
	return buf[:], nil
}

// NTPv4 structure
type NTPv4Packet struct {
	LeapIndicator     byte     //跳跃指示器（LeapIndicator）：2bit，指示NTP协议运行的状态，分为正常、提前、延后和未知状态。
	Version           uint8    //NTP版本号：3bit，用来指示使用的NTP版本号
	Mode              uint8    //模式：3bit，用来指示NTP客户端和服务器之间的交互方式
	Stratum           uint8    //NTP服务器的级别：8bit，用来指示NTP服务器的级别
	PollInterval      uint8    //客户端向服务器查询时间的间隔：8bit，用来指示客户端向服务器发送请求的间隔
	Precision         uint8    //NTP服务器时间的精度：8bit，用来指示NTP服务器时间的精度
	RootDelay         uint32   //根延迟：32bit，用来指示NTP客户端和服务器之间的延迟
	RootDisp          uint32   //根分散：32bit，用来指示NTP客户端和服务器之间的时间分散
	ReferenceID       uint32   //参考标识符：32bit，用来指示参考时钟的标识符
	RefTimestamp      int64    //参考时间戳：64bit，用来指示服务器的参考时间
	OrigTimestamp     int64    //发出时间戳：64bit，用来指示客户端发出请求的时间
	RecvTimestamp     int64    //接收时间戳：64bit，用来指示服务器接收请求的时间
	TransmitTimestamp int64    //发送时间戳：64bit，用来指示服务器发送响应的时间
	Auth              [8]uint8 //Authentication字段：64bit，用来验证报文的可靠性
}

// 判断是否为标准NTP请求报文
func IsStandardNtpRequest(pkt []byte) bool {
	return true
	//return (pkt[1] == 3) //pkt[0]=219 二进制为11011011
}

// 判断是否为Microsoft NTP请求报文
func IsMicrosoftNtpRequest(pkt []byte) bool {
	return (pkt[2] == 7)
}
