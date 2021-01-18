/**
 * @Time : 2021/1/11 7:47 下午
 * @Author : MassAdobe
 * @Description: utils
**/
package utils

import (
	"errors"
	"fmt"
	"math"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 3:35 下午
 * @Description: 常量池
**/
const (
	TimeFormatMS    = "2006-01-02 15:04:05"
	TimeFormatMonth = "2006-01-02"
)

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 3:36 下午
 * @Description: 运行当前系统命令
**/
func RunInLinuxWithErr(cmd string) (string, error) {
	result, err := exec.Command(cmd).Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	return strings.TrimSpace(string(result)), err
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 4:54 下午
 * @Description: 返回当前时间戳
**/
func RtnCurTime() string {
	return time.Now().Format(TimeFormatMS)
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 4:33 下午
 * @Description: 获取当前系统IP
**/
func ExternalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("没链接网络")
}

func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}
	return ip
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/13 6:02 下午
 * @Description: 校验IP地址是否正确
**/
func CheckIp(ip string) bool {
	addr := strings.Trim(ip, " ")
	regStr := `^(([1-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.)(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){2}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`
	if match, _ := regexp.MatchString(regStr, addr); match {
		return true
	}
	return false
}

/**
 * @Author: MassAdobe
 * @TIME: 2020-04-26 21:13
 * @Description: 返回时间字符串
**/
func RtnTmString() (timsStr string) {
	timsStr = time.Now().Format(TimeFormatMS)
	return
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/15 5:34 下午
 * @Description: 获取与当前相差时间(天)
**/
func GetHourDiffer(startTm string) (day int64) {
	day = math.MaxInt64
	t1, _ := time.ParseInLocation(TimeFormatMS, startTm, time.Local)
	t2 := time.Now()
	if t1.Before(t2) {
		day = (t2.Unix() - t1.Unix()) / (60 * 60 * 24)
	}
	return
}
