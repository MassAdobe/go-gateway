/**
 * @Time : 2021/1/11 7:53 下午
 * @Author : MassAdobe
 * @Description: go_gateway
**/
package main

import (
	"context"
	"fmt"
	"github.com/MassAdobe/go-gateway/constants"
	"github.com/MassAdobe/go-gateway/functions"
	"github.com/MassAdobe/go-gateway/logs"
	"github.com/MassAdobe/go-gateway/nacos"
	"github.com/MassAdobe/go-gateway/pojo"
	"github.com/MassAdobe/go-gateway/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/11 8:04 下午
 * @Description: 预热项
**/
func init() {
	fmt.Println(fmt.Sprintf(`{"log_level":"INFO","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "启动", "未知", "启动中"))
	s, _ := utils.RunInLinuxWithErr(constants.SYSTEM_CONTROL_PWD) // 执行linux命令获取当前路径
	sysData, _ := ioutil.ReadFile(s + constants.CONFIG_NAME)      // 读取系统配置
	if err := yaml.Unmarshal(sysData, &pojo.InitConf); err != nil {
		fmt.Println(fmt.Sprintf(`{"log_level":"INFO","time":"%s","msg":"%s","server_name":"%s","desc":"%s"}`, utils.RtnCurTime(), "启动", "未知", "解析系统配置失败"))
		os.Exit(1)
	}
	nacos.InitNacos()          // 初始化nacos配置
	nacos.NacosConfiguration() // nacos配置中心
	nacos.InitNacosProfile()   // 处理首次nacos获取到的配置信息
	logs.InitLogger(nacos.InitConfiguration.Log.Path,
		nacos.InitConfiguration.Serve.ServerName,
		nacos.InitConfiguration.Log.Level,
		nacos.InitConfiguration.Serve.Port) // 初始化日志
	nacos.ListenConfiguration()             // 监听配置文件变化
	nacos.NacosDiscovery()                  // nacos服务注册发现
	nacos.InitNacosGetInstances()           // 获取实例（初次获取）
}

/**
 * @Author: MassAdobe
 * @TIME: 2021/1/11 8:04 下午
 * @Description: 启动项
**/
func main() {
	server := &http.Server{ // 创建服务
		Addr:           constants.COLON_MARK + strconv.Itoa(int(nacos.InitConfiguration.Serve.Port)),
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        functions.NewMultipleHostsReverseProxy(),
	}
	logs.Lg.Info("启动", logs.Desc(fmt.Sprintf("启动端口: %d", nacos.InitConfiguration.Serve.Port)))
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed { // 监听并启动服务
			logs.Lg.Error("启动失败", err)
			os.Exit(1)
		}
	}()
	gracefulShutdown(server) // 优雅停服
}

/**
 * @Author: MassAdobe
 * @TIME: 2020/12/17 5:39 下午
 * @Description: 优雅停服
**/
func gracefulShutdown(server *http.Server) {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGSTOP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT, os.Interrupt)
	sig := <-quit
	logs.Lg.Info("准备关闭", logs.SpecDesc("收到信号", sig))
	now := time.Now()
	cxt, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	if err := server.Shutdown(cxt); err != nil {
		logs.Lg.Error("关闭失败", err)
	}
	nacos.NacosDeregister() // nacos注销服务
	logs.Lg.Info("退出成功", logs.Desc(fmt.Sprintf("退出花费时间: %v", time.Since(now))))
}
