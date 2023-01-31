package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"time"
)

const (
	NTP_PORT        = 123
	NTP_PACKET_SIZE = 48
	NTP_OFFSET      = 2208988800
)

type NTPPacket struct {
	LiVnMode       uint8 // 8 bits
	Stratum        uint8
	Poll           uint8
	Precision      uint8
	RootDelay      uint32
	RootDispersion uint32
	ReferenceID    uint32
	ReferenceTime  uint64
	OriginTime     uint64
	ReceiveTime    uint64
	TransmitTime   uint64
}

func main() {
	// 监听UDP端口
	socket, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: NTP_PORT,
	})
	if err != nil {
		fmt.Println("Failed to listen on UDP port", err)
		return
	}
	defer socket.Close()

	// 等待请求
	for {
		data := make([]byte, NTP_PACKET_SIZE)
		_, remoteAddr, err := socket.ReadFromUDP(data)
		fmt.Printf("接收到Packet remoteAddr:%s", remoteAddr)
		if err != nil {
			fmt.Println("Failed to receive data", err)
			continue
		}
		// 解析请求
		fmt.Println("解析接收到的Packet")
		packet := NTPPacket{}
		binary.Read(bufio.NewReader(bytes.NewReader(data)), binary.BigEndian, &packet)
		// 计算当前时间
		//currentTime := time.Now().UnixNano()
		//currentTime := time.Now().UTC().UnixNano()
		//strtimgnow := time.Now().UTC().UnixNano()
		//currentTime := GenNTPTimestamp()
		// convert to NTPv3/4 time
		ntpTime := GenNTPTimestamp()
		//currentTime := uint64(time.Now().UnixNano() + 2208988800000)
		//currentTime := time.Now().UTC().UnixNano() / 1e6 * int64(time.Millisecond)
		fmt.Println("解析Client.Packet:", packet)
		fmt.Printf("解析Client.Packet 16进制:%x\r\n", packet)
		// 更新NTP包信息
		fmt.Println("更新前packet.LiVnMode:", packet.LiVnMode)
		packet.LiVnMode = packet.LiVnMode & 0b11111000
		packet.LiVnMode = packet.LiVnMode | 0b00000100
		fmt.Println("更新后packet.LiVnMode:", packet.LiVnMode)
		packet.Stratum = 0
		packet.Poll = 10
		packet.Precision = 0
		//packet.RootDelay = uint32(uint64(ntpTime/1e7) - packet.ReferenceTime/1e7)
		packet.RootDelay = 0
		packet.RootDispersion = 0
		packet.ReferenceID = 0
		packet.ReferenceTime = ntpTime
		fmt.Println("更新后packet.ReferenceTime:", ntpTime)
		packet.OriginTime = 0
		packet.ReceiveTime = uint64(ntpTime)
		fmt.Println("更新前packet.TransmitTime:", packet.TransmitTime)
		packet.TransmitTime = uint64(ntpTime)
		//packet.TransmitTime = uint64(uint64(currentTime) - packet.TransmitTime)
		fmt.Println("更新后packet.TransmitTime:", ntpTime)
		//newpacket := packet

		// 将响应写入NTP包
		//bufio.NewWriter(bytes.NewBuffer(data)).Flush()
		//binary.Write(bufio.NewWriter(bytes.NewBuffer(data)), binary.BigEndian, packet)
		// 将NTP包结构体写入byte数组
		newdata := make([]byte, NTP_PACKET_SIZE)

		newdata[0] = packet.LiVnMode
		newdata[1] = packet.Stratum
		newdata[2] = packet.Poll
		newdata[3] = packet.Precision
		binary.BigEndian.PutUint32(newdata[4:8], packet.RootDelay)
		binary.BigEndian.PutUint32(newdata[8:12], packet.RootDispersion)
		binary.BigEndian.PutUint32(newdata[12:16], packet.ReferenceID)
		binary.BigEndian.PutUint64(newdata[16:24], packet.ReferenceTime)
		binary.BigEndian.PutUint64(newdata[24:32], packet.OriginTime)
		binary.BigEndian.PutUint64(newdata[32:40], packet.ReceiveTime)
		binary.BigEndian.PutUint64(newdata[40:48], packet.TransmitTime)
		// 向客户端发送响应
		fmt.Println("解析给Client.Packet:", newdata)
		fmt.Printf("解析给Client.Packet 16进制:%x", newdata)
		_, err = socket.WriteToUDP(newdata, remoteAddr)
		if err != nil {
			fmt.Println("Failed to send data", err)
			continue
		}
	}
}
func getNTPTime(time int64) uint32 {
	// NTP epoch is 1 Jan 1900
	epoch := int64(2208988800)
	// convert to NTP time
	ntpTime := uint32(math.Floor(float64(time-epoch) / float64(1000000000)))
	return ntpTime
}

// GenNTPTimestamp 生成20位长度的NTP时间戳
func GenNTPTimestamp() uint64 {
	//获取当前时间戳
	//now := time.Now().UnixNano()
	//将当前时间戳转换为NTP时间格式，即将时间戳的低32位放到高32位，将高32位放到低32位
	//nptTime := (uint64(now) << 32) | (uint64(now) >> 32)
	//return nptTime

	//1. 获取当前时间戳：
	//timestamp := time.Now().UnixNano()
	//2. 计算1900年1月1日0时0分0秒距离1970年1月1日0时0分0秒的总秒数：
	//start := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	//3. 将当前时间戳减去1900年1月1日0时0分0秒距离1970年1月1日0时0分0秒的总秒数，即可得到当前时间戳转换为NTP时间格式的结果：

	//ntpTime := uint64(timestamp - start)
	//ntpTime = ntpTime / 1000000000
	now := time.Now().UTC().UnixMilli() * 1e7
	//now := time.Now()
	//ntpTime := (now.Unix() + 2208988800) * 10000000
	return uint64(now)
	//1675130926791925000

}
