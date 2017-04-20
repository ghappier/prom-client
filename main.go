package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	//"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"

	"github.com/prometheus/prometheus/storage/remote"
)

var (
	Receiver = flag.String("receiver", "http://127.0.0.1:9990/receive", "接收数据的地址")
	Channels = flag.Int("channels", 1, "每批数据发送时的协程数")
	Times    = flag.Int("times", 1, "每个协程发送的数据量")
	Numps    = flag.Int("numps", 1, "每个协程发送多少数据量后睡眠1秒钟")
	Minutes  = flag.Int("minutes", 1, "发送多少批数据")
	Data     *string

	Client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	MetricNameLable = "__name__"
	MetricNames     = [...]string{
		"tomcat_cup_usage_persent",
		"tomcat_cup_idle_persent",
		/*
			"tomcat_menory_usage_persent",
			"tomcat_menory_idle_persent",
			"tomcat_disk_usage_persent",
			"tomcat_disk_idle_persent",
			"tomcat_io_usage_persent",
			"tomcat_io_idle_persent",
			"tomcat_network_usage_persent",
			"tomcat_network_idle_persent",
		*/

		"node_cup_usage_persent",
		"node_cup_idle_persent",
		/*
			"node_menory_usage_persent",
			"node_menory_idle_persent",
			"node_disk_usage_persent",
			"node_disk_idle_persent",
			"node_io_usage_persent",
			"node_io_idle_persent",
			"node_network_usage_persent",
			"node_network_idle_persent",
		*/
		/*
			"was_cup_usage_persent",
			"was_cup_idle_persent",
			"was_menory_usage_persent",
			"was_menory_idle_persent",
			"was_disk_usage_persent",
			"was_disk_idle_persent",
			"was_io_usage_persent",
			"was_io_idle_persent",
			"was_network_usage_persent",
			"was_network_idle_persent",

			"container_cup_usage_persent",
			"container_cup_idle_persent",
			"container_menory_usage_persent",
			"container_menory_idle_persent",
			"container_disk_usage_persent",
			"container_disk_idle_persent",
			"container_io_usage_persent",
			"container_io_idle_persent",
			"container_network_usage_persent",
			"container_network_idle_persent",

			"consul_cup_usage_persent",
			"consul_cup_idle_persent",
			"consul_menory_usage_persent",
			"consul_menory_idle_persent",
			"consul_disk_usage_persent",
			"consul_disk_idle_persent",
			"consul_io_usage_persent",
			"consul_io_idle_persent",
			"consul_network_usage_persent",
			"consul_network_idle_persent",

			"mq_cup_usage_persent",
			"mq_cup_idle_persent",
			"mq_menory_usage_persent",
			"mq_menory_idle_persent",
			"mq_disk_usage_persent",
			"mq_disk_idle_persent",
			"mq_io_usage_persent",
			"mq_io_idle_persent",
			"mq_network_usage_persent",
			"mq_network_idle_persent",
		*/
	}
	MetricNamesLen = len(MetricNames)

	HostnameLabel = "hostname"
	Hostnames     = [...]string{
		"dgg",
		"nkg",
		"szx",
	}
	HostnamesLen = len(Hostnames)

	DomainLabel = "domain"
	Domains     = [...]string{
		"dgg",
		//"nkg",
		//"szx",
	}
	DomainsLen = len(Domains)

	EnvLocationLabel = "env_location"
	EnvLocations     = [...]string{
		"dgg",
		//"nkg",
		//"szx",
	}
	EnvLocationsLen = len(EnvLocations)

	QualityLabel = "quality"
	Qualities    = [...]string{
		//"0",
		"0.25",
		//"0.75",
		//"1.0",
	}
	QualitiesLen = len(Qualities)

	ContainerIdLabel = "container"
	ContainerIds     = [...]string{
		"container_a",
		//"container_b",
		//"container_c",
	}
	ContainerIdLen = len(ContainerIds)

	Tmpl string = `
{
	"timeseries":[
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]},
		{"labels":[{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"},{"name":"%s","value":"%s"}],"samples":[{"value":%f,"timestamp_ms":%d}]}
	]
}
	`
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if flag.NArg() == 0 {
		flag.PrintDefaults()
		//os.Exit(0)
	}
	log.Println(time.Now(), " : 开始发送数据")
	begin := time.Now()

	for i := 0; i < *Minutes; i++ {
		go sendBatch(i + 1)
		time.Sleep(1 * time.Minute)
	}
	complete := time.Now()
	log.Println(time.Now(), " : 所有数据发送完成，耗时：", complete.Sub(begin))
}

func sendBatch(index int) {
	log.Println(time.Now(), " : 开始发送第", index, "批数据")
	start := time.Now()
	wg := new(sync.WaitGroup)
	for i := 0; i < *Channels; i++ {
		wg.Add(1)
		go send(wg)
	}
	wg.Wait()
	end := time.Now()
	log.Println(time.Now(), " : 第", index, "批数据发送完成，耗时：", end.Sub(start))
}

func send(wg *sync.WaitGroup) {
	defer wg.Done()
	var req remote.WriteRequest
	spider, err := NewSpider()
	if err != nil {
		log.Fatalf("error :%s\n", err.Error())
	} else {
		spider.Method = "post"
		spider.Url = *Receiver
		for i := 0; i < *Times; i++ {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			data := fmt.Sprintf(Tmpl,
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
				MetricNameLable, MetricNames[r.Intn(MetricNamesLen)], HostnameLabel, Hostnames[r.Intn(HostnamesLen)]+strconv.Itoa(2), DomainLabel, Domains[r.Intn(DomainsLen)]+strconv.Itoa(1), EnvLocationLabel, EnvLocations[r.Intn(EnvLocationsLen)]+strconv.Itoa(1), QualityLabel, Qualities[r.Intn(QualitiesLen)], ContainerIdLabel, ContainerIds[r.Intn(ContainerIdLen)], r.Float32(), time.Now().Unix(),
			)
			if err := json.Unmarshal([]byte(data), &req); err != nil {
				log.Printf("json unmarshal error : %s\n", err.Error())
				continue
			}
			pdata, err := proto.Marshal((proto.Message)(&req))
			if err != nil {
				log.Printf("proto unmarshal error : %s\n", err.Error())
				continue
			}
			buf := bytes.Buffer{}
			if _, err := snappy.NewWriter(&buf).Write(pdata); err != nil {
				log.Printf("snappy write error : %s\n", err.Error())
				continue
			}
			spider.Data = buf
			if _, err := spider.PostBody(); err != nil {
				log.Printf("response error from server : %s\n", err.Error())
			}
			if i%*Numps == 0 {
				time.Sleep(1 * time.Second)
			}
		}
	}
}

type Spider struct {
	Url           string
	UrlStatusCode int
	Method        string
	Data          bytes.Buffer
	Client        *http.Client
}

func NewClient() (*http.Client, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
	return client, nil
}

func NewSpider() (*Spider, error) {
	spider := new(Spider)
	client, err := NewClient()
	spider.Client = client
	return spider, err
}

func (this *Spider) PostBody() (body []byte, e error) {
	var request = &http.Request{}
	request, _ = http.NewRequest("POST", this.Url, &this.Data)
	request.Header.Add("Content-Encoding", "snappy")
	if this.Client == nil {
		this.Client = Client
	}
	response, err := this.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	this.UrlStatusCode = response.StatusCode
	body, e = ioutil.ReadAll(response.Body)
	return
}
