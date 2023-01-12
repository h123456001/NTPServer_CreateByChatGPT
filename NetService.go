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
	pkt.Stratum = buf[1]                                  //NTP服务器的级别：8bit，用来指示NTP服务器的级别
	pkt.PollInterval = buf[2]                             //客户端向服务器查询时间的间隔：8bit，用来指示客户端向服务器发送请求的间隔
	pkt.Precision = buf[3]                                ////NTP服务器时间的精度：8bit，用来指示NTP服务器时间的精度 1.0seconed=0b000000
	pkt.RootDelay = binary.BigEndian.Uint32(buf[4:8])     //根延迟：4byte=32bit，用来指示NTP客户端和服务器之间的延迟
	pkt.RootDisp = binary.BigEndian.Uint32(buf[8:12])     //根分散：32bit，用来指示NTP客户端和服务器之间的时间分散
	pkt.ReferenceID = binary.BigEndian.Uint32(buf[12:16]) //参考标识符：32bit，用来指示参考时钟的标识符https://www.rfc-editor.org/rfc/rfc5905 搜索关键字"Reference ID" 3/17
	//pkt.RefTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[16:24])-2208988800), 0).Unix()      //参考时间戳：64bit，用来指示服务器的参考时间
	clentReftime := binary.BigEndian.Uint64(buf[16:24]) //客户端的unix时间戳 utc,
	fmt.Println(clentReftime)
	fmt.Println("client UnixTime=", time.Unix(int64(clentReftime), 0).UTC(), "TimeUnixMicro=", time.UnixMicro(int64(clentReftime)).UTC())
	//从buf中解析出RefTimeStamp
	refTimeStamp := binary.BigEndian.Uint32(buf[40:44])
	//将RefTimeStamp转换成时间格式
	fmt.Println("RefTimeStamp：", refTimeStamp)
	fmt.Println("RefTimeStamp标准时间格式：", time.Unix(int64(refTimeStamp-2208988800), 0)) //----这里时间就对了是客户端现在的时间----
	pkt.RefTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[16:24])), 0).Unix()
	pkt.OrigTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[24:32])-2208988800), 0).Unix()     //发出时间戳：64bit，用来指示客户端发出请求的时间
	pkt.RecvTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[32:40])-2208988800), 0).Unix()     //接收时间戳：64bit，用来指示服务器接收请求的时间
	pkt.TransmitTimestamp = time.Unix(int64(binary.BigEndian.Uint32(buf[40:48])-2208988800), 0).Unix() //发送时间戳：64bit，用来指示服务器发送响应的时间
	return pkt, nil
}

func CreateNTPResponse(pkt NTPv4Packet) ([]byte, error) {
	// Create response packet
	var buf [NtpV4PacketSize]byte
	/*
			LeapIndicator     byte     //跳跃指示器（LeapIndicator）：2bit，指示NTP协议运行的状态，分为正常、提前、延后和未知状态。	[0]aabbbccc中的aa
			Version           uint8    //NTP版本号：3bit，用来指示使用的NTP版本号				[0]aabbbccc中的bbb
			Mode              uint8    //模式：3bit，用来指示NTP客户端和服务器之间的交互方式	[0]aabbbccc中的ccc
			Stratum           uint8    //NTP服务器的级别：8bit，用来指示NTP服务器的级别		[1]
			PollInterval      uint8    //客户端向服务器查询时间的间隔：8bit，用来指示客户端向服务器发送请求的间隔	[2]
			Precision         int8     //NTP服务器时间的精度：8bit，用来指示NTP服务器时间的精度	[3]		服务端自定义为0
			RootDelay         int32    //根延迟：32bit，用来指示NTP客户端和服务器之间的延迟		[4-8]	服务端自定义为0
			RootDisp          int32    //根分散：32bit，用来指示NTP客户端和服务器之间的时间分散	[8-12]	服务端自定义为0
			ReferenceID       uint32   //参考标识符：32bit，用来指示参考时钟的标识符			[12-16]	服务端自定义为服务端ip
			RefTimestamp      uint64   //参考时间戳：64bit，用来指示服务器的参考时间			[16-24]	服务端时间time.Now().UTC().Unix()
			OrigTimestamp     uint64   //发出时间戳：64bit，用来指示客户端发出请求的时间		[24-32] 客户端时间client_pkt.Transmit
			RecvTimestamp     uint64   //接收时间戳：64bit，用来指示服务器接收请求的时间		[32-40]	服务端时间time.Now().UTC().Unix()
			TransmitTimestamp uint64   //发送时间戳：64bit，用来指示服务器发送响应的时间		[40-48]	服务端时间time.Now().UTC().Unix()
			Auth              [8]uint8 //Authentication字段：64bit，用来验证报文的可靠性
		是的，客户端发出请求的时间戳对应的值是Payload里的Reference Timestamp，它对于服务端来说是毫无意义的，因为服务端需要计算客户端发送请求的时间戳，而不是使用客户端发送的时间戳。
		服务端（服务的IP为192.168.16.120）需要解析返回给客户端的Payload里包括("->"你可以理解我服务端对对应属性的自定义赋值)：
		Mode->4
		Stature->0
		Root Delay->0
		Root Dispersion->0
		Reference Identifier->192.168.16.120
		Reference Timestamp->time.Now().UTC().Unix()
		Originate Timestamp->clientpkt.Transmit
		Receive Timestamp->time.Now().UTC().Unix()
		Transmit Timestamp->time.Now().UTC().Unix()
		我的理解是否正确？
		是的，您的理解是正确的。Mode替换为4，Stature替换为0，Root Delay替换为0，Root Dispersion替换为0，Reference Identifier替换为服务端的IP地址，Reference Timestamp替换为服务端当前的UTC时间戳，Originate Timestamp替换为客户端发送的Transmit Timestamp，Receive Timestamp替换为服务端当前的UTC时间戳，Transmit Timestamp替换为服务端当前的UTC时间戳。
	*/
	//buf[0]最后3位 客户端3 服务端4  client 00 100 011 server 00 100
	var ln uint8 = 0b00000000 //pkt.LeapIndicator 服务端要给客户端00 表示时间是准确的除非服务端有其他服务器可以同步
	ver := uint8(4)           //pkt.Version
	mode := pkt.Mode + 0b00000001
	result := (ln << 6) | (ver << 3) | mode
	buf[0] = result
	buf[0] = buf[0] & 0b11111100
	//var test byte = 0b11111100
	fmt.Println("buf[0]:%v", buf[0])
	//buf[0] = (pkt.LeapIndicator << 6) | (pkt.Version << 3) | pkt.Mode
	buf[1] = 0b00000011                //pkt.Stratum //服务器时间级别自定义为3 0b00000011=3  来时其他 2 1 级同步结果
	buf[2] = 0b00000000                //pkt.PollInterval 可以设置为0对服务端而言 该值无意义
	buf[3] = 0b00000000                //pkt.Precision NTP服务器时间的精度：8bit，用来指示NTP服务器时间的精度
	copy(buf[4:8], []byte{0, 0, 0, 0}) //[4-8]	服务端自定义为0 RootDelay-int32	根延迟：32bit，用来指示NTP客户端和服务器之间的延迟
	//binary.BigEndian.PutUint32(buf[4:8], uint32(pkt.RootDelay)) //根延迟：32bit，用来指示NTP客户端和服务器之间的延迟 RootDelay值 不包含网络传输所需时间，而是由NTP服务器从四个精确的NTP服务器获取标准时间，并计算出与服务器的延迟，然后根据计算出的延迟值计算出RootDelay的值。---RootDelay值取决于NTP服务器之间的网络延迟，如果没有其他NTP服务器可参考，可以将RootDelay值设置为0或较小的值。

	copy(buf[8:12], []byte{0, 0, 0, 0}) //[8-12]	服务端自定义为0RootDisp-int32根分散：32bit，用来指示NTP客户端和服务器之间的时间分散
	//binary.BigEndian.PutUint32(buf[8:12], uint32(pkt.RootDisp)) //RootDisp值是根据RootDelay值计算出来的，如果RootDelay值设置为0或较小的值，那么RootDisp值也会很小。

	copy(buf[12:16], net.IPv4(192, 168, 16, 120)) //[12-16]	服务端自定义为服务端ip ReferenceID-uint32	参考标识符：32bit，用来指示参考时钟的标识符
	//binary.BigEndian.PutUint32(buf[12:16], pkt.ReferenceID)//一般为服务器ipv4
	servertime := uint64(time.Now().Unix() - 2208988800)
	//copy(buf[16:24], string(time.Now().UTC().Unix())) //[16-24]	服务端时间time.Now().UTC().Unix() RefTimestamp-uint64	参考时间戳：64bit，用来指示服务器的参考时间

	binary.BigEndian.PutUint64(buf[16:24], servertime)               //[16-24]	服务端时间time.Now().UTC().Unix() RefTimestamp-uint64	参考时间戳：64bit，用来指示服务器的参考时间
	binary.BigEndian.PutUint64(buf[24:32], uint64(pkt.RefTimestamp)) //[24-32] 客户端时间client_pkt.Transmit	OrigTimestamp-uint64	发出时间戳：64bit，用来指示客户端发出请求的时间
	binary.BigEndian.PutUint64(buf[32:40], servertime)               //[32-40]	服务端时间time.Now().UTC().Unix()	RecvTimestamp-uint64	接收时间戳：64bit，用来指示服务器接收请求的时间

	//binary.BigEndian.PutUint32(buf[40:48], uint32(transmitTime))
	binary.BigEndian.PutUint64(buf[40:48], servertime) //[40-48]	服务端时间time.Now().UTC().Unix()	TransmitTimestamp-uint64	发送时间戳：64bit，用来指示服务器发送响应的时间
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
