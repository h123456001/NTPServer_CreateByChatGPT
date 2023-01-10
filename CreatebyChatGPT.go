package main

/*那用Go语言开发NTP服务端 读取配置文件ntpserver.conf，配置文件内有：
#上级NTP服务器的IP地址 不填写表示本地实际
ntpserverip:10.10.10.10
#更新频率 秒  若本地时间不填写无意义
updatefrequency:3600
#配置文件结束
启动该程序自动读取配置文件自动填入读取的参数，若无参数则为空并提示长期保存请求改配置文件。并提示10秒内无操作自动继续
10秒后 启动自定义的NTP服务器。
配置文件内有值的情况下：先作为客户端请求远程NTP服务器时间更新本地时间。并获取作为NTP服务器需要填入的RootDelay、ReferenceID等参数 存在内存变量里。
配置文件内无值的情况下：直接作为NTP服务器 填入自定义的RootDelay、ReferenceID等参数  等待客户端来同步时间。

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

// 定义全局变量
var ntpServerIP string
var updateFrequency int

// 读取配置文件
func readConfig() {
	data, err := ioutil.ReadFile("ntpserver.conf")
	if err != nil {
		fmt.Println("Read config failed.")
		return
	}
	reader := bufio.NewReader(strings.NewReader(string(data)))
	// 读取上级NTP服务器IP
	ntpServerIP, _ = reader.ReadString('\n')
	// 读取更新频率
	updateFrequency, _ = reader.ReadString('\n')
	fmt.Println("Read config success.")
}

// 启动NTP服务
func startNTPServer() {
	// 启动UDP监听
	addr, err := net.ResolveUDPAddr("udp", ":123")
	if err != nil {
		fmt.Println("Resolve UDP address failed.")
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Listen UDP failed.")
		return
	}
	defer conn.Close()
	// 启动循环更新
	go loopUpdate(conn)
	// 如果有上级NTP服务器，则先作为客户端请求服务器时间
	if ntpServerIP != "" {
		// 根据配置的上级NTP服务器IP地址，构造请求地址
		ntpserverAddr, err := net.ResolveUDPAddr("udp", ntpServerIP)
		if err != nil {
			fmt.Println("Resolve NTP server address failed.")
			return
		}
		// 向上级NTP服务器发送请求
		data := make([]byte, 48)
		_, err = conn.WriteToUDP(data, ntpserverAddr)
		if err != nil {
			fmt.Println("Send request to NTP server failed.")
			return
		}
		// 接收上级NTP服务器响应，同步本地时间
		_, _, err = conn.ReadFromUDP(data)
		if err != nil {
			fmt.Println("Receive response from NTP server failed.")
			return
		}
		// 根据响应内容解析出时间
		sec := uint64(data[43]) | uint64(data[42])<<8 | uint64(data[41])<<16 | uint64(data[40])<<24
		frac := uint64(data[47]) | uint64(data[46])<<8 | uint64(data[45])<<16 | uint64(data[44])<<24
		nsec := sec*1e9 + (frac*1e9)>>32
		// 设置本地时间
		t := time.Unix(int64(sec)-2208988800, int64(nsec))
		time.LoadLocation("Local")
		time.Local = t
		fmt.Println("Sync time success.")
	}

	// 开始接收客户端请求
	for {
		// 接收客户端请求
		_, remoteAddr, err := conn.ReadFromUDP(data)
		if err != nil {
			fmt.Println("Receive request failed.")
			continue
		}
		// 获取本地时间
		now := time.Now()
		sec := uint32(now.Unix() + 2208988800)
		frac := uint32(now.Nanosecond() * 4294.967296)
		// 构造本地时间响应报文
		data[0] = 0x1b
		data[1] = 0x00
		data[2] = 0x00
		data[3] = 0x00
		data[4] = 0x00
		data[5] = 0x00
		data[6] = 0x00
		data[7] = 0x00
		data[8] = 0x00
		data[9] = 0x00
		data[10] = 0x00
		data[11] = 0x00
		data[12] = 0x00
		data[13] = 0x00
		data[14] = 0x00
		data[15] = 0x00
		data[16] = 0x00
		data[17] = 0x00
		data[18] = 0x00
		data[19] = 0x00
		data[20] = 0x00
		data[21] = 0x00
		data[22] = 0x00
		data[23] = 0x00
		data[24] = byte(sec >> 24)
		data[25] = byte(sec >> 16)
		data[26] = byte(sec >> 8)
		data[27] = byte(sec)
		data[28] = byte(frac >> 24)
		data[29] = byte(frac >> 16)
		data[30] = byte(frac >> 8)
		data[31] = byte(frac)
		// 根据配置填充RootDelay、ReferenceID等参数
		data[32] = 0x00
		data[33] = 0x00
		data[34] = 0x00
		data[35] = 0x00
		data[36] = 0x00
		data[37] = 0x00
		data[38] = 0x00
		data[39] = 0x00
		data[40] = 0x00
		data[41] = 0x00
		data[42] = 0x00
		data[43] = 0x00
		data[44] = 0x00
		data[45] = 0x00
		data[46] = 0x00
		data[47] = 0x00
		// 将响应报文发送给客户端
		_, err = conn.WriteToUDP(data, remoteAddr)
		if err != nil {
			fmt.Println("Send response to client failed.")
			continue
		}
		fmt.Println("Send response to client success.")
	}
}

// 循环更新本地时间
func loopUpdate(conn *net.UDPConn) {
	for {
		// 如果有上级NTP服务器，则向上级NTP服务器发送请求
		if ntpServerIP != "" {
			// 根据配置的上级NTP服务器IP地址，构造请求地址
			ntpserverAddr, err := net.ResolveUDPAddr("udp", ntpServerIP)
			if err != nil {
				fmt.Println("Resolve NTP server address failed.")
				return
			}
			// 向上级NTP服务器发送请求
			data := make([]byte, 48)
			_, err = conn.WriteToUDP(data, ntpserverAddr)
			if err != nil {
				fmt.Println("Send request to NTP server failed.")
				return
			}
			// 接收上级NTP服务器响应，同步本地时间
			_, _, err = conn.ReadFromUDP(data)
			if err != nil {
				fmt.Println("Receive response from NTP server failed.")
				return
			}
			// 根据响应内容解析出时间
			sec := uint64(data[43]) | uint64(data[42])<<8 | uint64(data[41])<<16 | uint64(data[40])<<24
			frac := uint64(data[47]) | uint64(data[46])<<8 | uint64(data[45])<<16 | uint64(data[44])<<24
			nsec := sec*1e9 + (frac*1e9)>>32
			// 设置本地时间
			t := time.Unix(int64(sec)-2208988800, int64(nsec))
			time.LoadLocation("Local")
			time.Local = t
			fmt.Println("Sync time success.")
		}
		// 根据配置文件里的更新频率，每次更新完后等待指定时间
		time.Sleep(time.Second * time.Duration(updateFrequency))
	}
}

func main() {
	// 读取配置文件
	readConfig()
	// 启动NTP服务
	startNTPServer()
}
*/
