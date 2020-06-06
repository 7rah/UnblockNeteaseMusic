package main

import (
	"fmt"

	"github.com/7rah/UnblockNeteaseMusic/config"

	//_ "github.com/mkevac/debugcharts" // 可选，添加后可以查看几个实时图表数据

	"github.com/7rah/UnblockNeteaseMusic/proxy"
	//_ "net/http/pprof" // 必须，引入 pprof 模块
)

func main() {
	fmt.Println("--------------------Config--------------------")
	fmt.Println("port     =", config.Port)
	fmt.Println("source   =", config.Source)
	fmt.Println("mode     =", config.Mode)
	fmt.Println("endPoint =", config.EndPoint)
	proxy.InitProxy()

}
