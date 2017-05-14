package main

import (
	"flag"
	"fmt"
	"net/url"
	"runtime"
	"sync"
	"time"

	"github.com/ghappier/prom-client/client"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/config"
)

var (
	Receiver  = flag.String("receiver", "http://127.0.0.1:9990/receive", "接收数据的地址，格式：http://127.0.0.1:9990/receive")
	MetricNum = flag.Int("metric-num", 100, "每台机器的指标数量")
	VmNum     = flag.Int("vm-num", 50, "机器的数量")
	Numps     = flag.Int("numps", 500, "发送多少批数据后睡眠1秒钟")
	Batchs    = flag.Int("batchs", 5, "发送多少批数据")
	Interval  = flag.Int("interval", 60000, "间隔多少毫秒发送一批数据")
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if flag.NArg() == 0 {
		flag.PrintDefaults()
	}
	metrics, hosts, err := client.InitData(*MetricNum, *VmNum)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var index int = 1
	var conf *client.ClientConfig = new(client.ClientConfig)
	murl, merr := url.Parse(*Receiver)
	if merr != nil {
		fmt.Printf("url parse error:%s\n", merr.Error())
	}
	conf.Url = &config.URL{murl}
	conf.Timeout = model.Duration(5 * time.Minute)
	mclient, err := client.NewClient(index, conf)
	if err != nil {
		fmt.Printf("new client error, %s", err.Error())
		return
	}
	fmt.Println(time.Now(), " : 开始发送数据")
	begin := time.Now()
	wgroup := new(sync.WaitGroup)
	for i := 0; i < *Batchs; i++ {
		wgroup.Add(1)
		go func(index int, total int, wg *sync.WaitGroup) {
			defer wg.Done()
			start := time.Now()
			fmt.Println(time.Now(), " : 开始发送第", index, "/", total, "批数据")
			samples := client.GenSamples(metrics, hosts)
			client.Send(mclient, samples, *Numps)
			end := time.Now()
			fmt.Println(time.Now(), " : 第", index, "批数据发送完成，耗时：", end.Sub(start))
		}(i+1, *Batchs, wgroup)
		if i < *Batchs-1 {
			time.Sleep(time.Duration(*Interval) * time.Millisecond)
		}
	}
	wgroup.Wait()
	complete := time.Now()
	fmt.Println(time.Now(), " : 所有数据发送完成，耗时：", complete.Sub(begin))

}
