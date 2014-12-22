# GoFlow - Dataflow and Flow-based programming library for Go (golang)

This is quite a minimalistic implementation of [Flow-based programming](http://en.wikipedia.org/wiki/Flow-based_programming) and several other concurrent models in Go programming language that aims at designing applications as graphs of components which react to data that flows through the graph.

The main properties of the proposed model are:

* Concurrent - graph nodes run in parallel.
* Structural - applications are described as components, their ports and connections between them.
* Event-driven - system's behavior is how components react to events.
* Asynchronous - there is no determined order in which events happen.
* Isolated - sharing is done by communication, state is not shared.

## Getting started

Current version of the library requires a latest stable Go release. If you don't have the Go compiler installed, read the official [Go install guide](http://golang.org/doc/install).

Use go tool to install the package in your packages tree:

```
go get github.com/gogap/flow
```

Then you can use it in import section of your Go programs:

```go
import "github.com/gogap/flow"
```

## Basic Example

Below there is a listing of a simple program running a network of two processes.

![Greeter example diagram](http://flowbased.wdfiles.com/local--files/goflow/goflow-hello.png)

This first one generates greetings for given names, the second one prints them on screen. It demonstrates how components and graphs are defined and how they are embedded into the main program.

```go
package main

import (
	"net/http"
	"time"

	"github.com/gogap/flow"
)

//APP,必须组合Graph
type GreetingApp struct {
	flow.Graph
}

//与外部世界交互的入口
type Greeter struct {
	flow.Component                 // 必须组合component
	Ipt            <-chan Message1 `flow:"input"`
	Opt            chan<- Message2 `flow:"sender"`
}

//接受数据并转发到Printer
func (g *Greeter) OnIpt(ipt Message1) {
	var ipt2 Message2
	ipt2.Output = ipt.Output //返回channel逐级向下传递
	ipt2.Time = ipt.Time
	g.Opt <- ipt2
}

//返回数据到外部世界
type Printer struct {
	flow.Component
	Ipt <-chan Message2 `flow:"receiver"`
}

//返回到Output
func (p *Printer) OnIpt(ipt2 Message2) {
	ipt2.Output <- ipt2.Time
}

//传递的消息之一，必须组合Context
type Message1 struct {
	flow.Context
	Time string
}

//传递的消息之二，必须组合Context
type Message2 struct {
	flow.Context
	Time string
}

func main() {
	net := flow.NewNet()                    //新建网络
	net.Add(new(Greeter), new(Printer))     //增加组件
	net.Connect(new(Greeter), new(Printer)) //连接组件（发送方、接收方）
	net.Run()

	in := make(chan Message1)       //初始化对象
	net.SetInPort(new(Greeter), in) //映射到网络入口

	var f = func(writer http.ResponseWriter, req *http.Request) {
		ipt := Message1{Time: time.Now().String()}
		ipt.Context.Init()  //必须初始化channel，否则阻塞且不报错！
		in <- ipt           //数据进入网络
		opt := <-ipt.Output //网络返回数据
		if ipt.Time != opt.(string) {
			panic("mix channel")
		}
	}

	http.HandleFunc("/", f)
	err := http.ListenAndServe(":8000", nil) //模拟完整的web server
	if err != nil {
		panic(err)
	}
}
```

Looks a bit heavy for such a simple task but FBP is aimed at a bit more complex things than just printing on screen. So in more complex an realistic examples the infractructure pays the price.

You probably have one question left even after reading the comments in code: why do we need to wait for the finish signal? This is because flow-based world is asynchronous and while you expect things to happen in the same sequence as they are in main(), during runtime they don't necessarily follow the same order and the application might terminate before the network has done its job. To avoid this confusion we listen for a signal on network's `Wait()` channel which is closed when the network finishes its job.

## Terminology

Here are some Flow-based programming terms used in GoFlow:

* Component - the basic element that processes data. Its structure consists of input and output ports and state fields. Its behavior is the set of event handlers. In OOP terms Component is a Class.
* Connection - a link between 2 ports in the graph. In Go it is a channel of specific type.
* Graph - components and connections between them, forming a higher level entity. Graphs may represent composite components or entire applications. In OOP terms Graph is a Class.
* Network - is a Graph instance running in memory. In OOP terms a Network is an object of Graph class.
* Port - is a property of a Component or Graph through which it communicates with the outer world. There are input ports (Inports) and output ports (Outports). For GoFlow components it is a channel field.
* Process - is a Component instance running in memory. In OOP terms a Process is an object of Component class.

More terms can be found in [flowbased terms](http://flowbased.org/terms) and [FBP wiki](http://www.jpaulmorrison.com/cgi-bin/wiki.pl?action=index).

## Documentation

### Contents

1. [Components](https://github.com/trustmaster/goflow/wiki/Components)
    1. [Ports, Events and Handlers](https://github.com/trustmaster/goflow/wiki/Components#ports-events-and-handlers)
    2. [Processes and their lifetime](https://github.com/trustmaster/goflow/wiki/Components#processes-and-their-lifetime)
    3. [State](https://github.com/trustmaster/goflow/wiki/Components#state)
    4. [Concurrency](https://github.com/trustmaster/goflow/wiki/Components#concurrency)
    5. [Internal state and Thread-safety](https://github.com/trustmaster/goflow/wiki/Components#internal-state-and-thread-safety)
2. [Graphs](https://github.com/trustmaster/goflow/wiki/Graphs)
    1. [Structure definition](https://github.com/trustmaster/goflow/wiki/Graphs#structure-definition)
    2. [Behavior](https://github.com/trustmaster/goflow/wiki/Graphs#behavior)

### Package docs

Documentation for the flow package can be accessed using standard godoc tool, e.g.

```
godoc github.com/trustmaster/goflow
```

## More examples

* [GoChat](https://github.com/trustmaster/gochat), a simple chat in Go using this library

## Links

Here are related projects and resources:

* [J. Paul Morrison's Flow-Based Programming](http://www.jpaulmorrison.com/fbp/), the origin of FBP, [JavaFBP, C#FBP](http://sourceforge.net/projects/flow-based-pgmg/) and [DrawFBP](http://www.jpaulmorrison.com/fbp/#DrawFBP) diagramming tool.
* [Knol about FBP](http://knol.google.com/k/flow-based-programming)
* [NoFlo](http://noflojs.org/), FBP for JavaScript and Node.js
* [Pypes](http://www.pypes.org/), flow-based Python ETL
* [Go](http://golang.org/), the Go programming language

## TODO

* Integration with NoFlo-UI
* Distributed networks via TCP/IP and UDP
* Better run-time restructuring and evolution
* Reflection and monitoring of networks
