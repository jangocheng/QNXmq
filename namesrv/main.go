// Copyright 2018 liberxue

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//    http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
/**
* Broker与Namesrv的心跳机制
单个Broker跟所有Namesrv保持心跳请求，心跳间隔为30秒，心跳请求中包括当前Broker所有的Topic信息。
Namesrv会反查Broer的心跳信息，如果某个Broker在2分钟之内都没有心跳，则认为该Broker下线(主管下限和客观下限类似redis)，调整Topic
跟Broker的对应关系。但此时Namesrv不会主动通知Producer、Consumer有Broker宕机。
*
*
* 消费者启动时需要指定Namesrv地址，与其中一个Namesrv建立长连接。消费者每隔30秒从nameserver获取所有topic
* 的最新队列情况，这意味着某个broker如果宕机，客户端最多要30秒才能感知。连接建立后，从namesrv中获取当前
* 消费Topic所涉及的Broker，直连Broker。
* Consumer跟Broker是长连接，会每隔30秒发心跳信息到Broker。Broker端每10秒检查一次当前存活的Consumer，若发
* 现某个Consumer 2分钟内没有心跳，就断开与该Consumer的连接，并且向该消费组的其他实例发送通知，触发该消费者集群的负载均衡。
*
* Producer启动时，也需要指定Namesrv的地址，从Namesrv集群中选一台建立长连接。如果该Namesrv宕机，会自动连其他Namesrv。直到有可用的Namesrv为止。
* 生产者每30秒从Namesrv获取Topic跟Broker的映射关系，更新到本地内存中。再跟Topic涉及的所有Broker建立长连接，每隔30秒发一次心跳。
* 在Broker端也会每10秒扫描一次当前注册的Producer，如果发现某个Producer超过2分钟都没有发心跳，则断开连接。
**/
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/Liberxue/MQ/common"
	"github.com/smallnest/goframe"
)

func main() {
	// c := flag.String("c", "configFile", "Name server config properties file")
	p := flag.String("p", "printConfigItem", "Print all config item")
	v := flag.Bool("v", false, "version")
	flag.Parse()
	if *v {
		fmt.Println("MQ namesrv version:", common.Version)
		os.Exit(0)
	}
	l, err := net.Listen("tcp", ":9876")
	if err != nil {
		panic(err)
	}

	defer l.Close()

	encoderConfig := goframe.EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               4,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}

	decoderConfig := goframe.DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    0,
		InitialBytesToStrip: 4,
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		c := goframe.NewLengthFieldBasedFrameConn(encoderConfig, decoderConfig, conn)
		go func(conn goframe.FrameConn) {
			for {
				b, err := c.ReadFrame()
				if err != nil {
					if err == io.EOF {
						return
					}
					panic(err)
				}
				fmt.Println(string(b))

				s := fmt.Sprintf("%d: %s", time.Now().UnixNano()/1e6, string(b))
				c.WriteFrame([]byte(s))
			}
		}(c)
	}

	// * */  //nameserver依赖 RocketmqCommon 和 RocketmqRemoting，各个模块的日志级别如上配置
	//例如listenPort=9998是通过命令行携带的，则以命令行为准，即使加到配置文件中了，因为这里后执行，所以还是以命令行为准
}
