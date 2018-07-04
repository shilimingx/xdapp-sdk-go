package register

import (
	"github.com/alecthomas/log4go"
	"bytes"
)

const (
	defaultHostIsSSl   = true
	defaultServiceApp  = "test"
	defaultServiceName = "console"
	defaultHost        = "www.xdapp.com:8900"
)

type SRegister struct {
	console
	Logger    *log4go.Logger					// 创建的tcp客户端对象
	Client    *Client							// 创建的tcp客户端对象
	RegSuccess  bool							// 注册成功标志
	ServiceData (map[string]map[string]string)	// console 注册成功返回的页面服务器信息
}

var (
	MyRpc   = NewMyRpc()
	MyLog   *log4go.Logger // log日志
	isDebug = false // 是否debug模式
	logName = "test.log"
)

/**
	tcp 配置
 */
var tcpConf = &tcpConfig {
	1,				// 包长开始位
	13,				// 1字节消息类型+4字节消息体长度+4字节用户id+4字节原消息fd+内容（id+data）
	0x200000}			// 最大包长度

/**
	工厂创建
 */
func NewRegister() *SRegister {

	Console := conf.Console

	if Console.Host == "" {
		Console.Host = defaultHost
	}
	if Console.SSl == false {
		Console.SSl = defaultHostIsSSl
	}
	if Console.App == "" {
		Console.App = defaultServiceApp
	}
	if Console.Name == "" {
		Console.Name = defaultServiceName
	}

	MyLog = NewLog4go()
	client := NewClient(Console.Host, *tcpConf)

	return &SRegister{Console, MyLog,client, false, make (map[string]map[string]string)}
}

func (reg *SRegister) GetApp() string {
	return reg.App
}

func (reg *SRegister) GetName() string {
	return reg.Name
}
func (reg *SRegister) GetKey() string {
	return reg.Key
}

func (reg *SRegister) SetRegSuccess(status bool) {
	reg.RegSuccess = status
}

func (reg *SRegister) SetServiceData(data map[string]map[string]string) {
	reg.ServiceData = data
}

func (reg *SRegister) CloseClient() {
	reg.Client.Close(reg.RegSuccess)
}

func (reg *SRegister) Info(arg0 interface{}, args ...interface{}) {
	reg.Logger.Info(arg0, args ...)
}

func (reg *SRegister) Debug(arg0 interface{}, args ...interface{}) {
	reg.Logger.Debug(arg0, args ...)
}

func (reg *SRegister) Warn(arg0 interface{}, args ...interface{}) {
	reg.Logger.Warn(arg0, args ...)
}

func (reg *SRegister) Error(arg0 interface{}, args ...interface{}) {
	reg.Logger.Error(arg0, args ...)
}

/**
	tcp client
 */
func (reg *SRegister) CreateClient() {

	reg.Client.OnReceive(func(message []byte) {

		request := new(ReqestData)
		request.Unpack(bytes.NewReader(message))

		//myRpc.context.BaseContext.Set("receiveParam")

		// 执行rpc返回
		rpcData := MyRpc.handle(request.Data, MyRpc.context)

		rs := string(PackId(request.Id)) + string(rpcData)

		dataLen := len(rs);
		if dataLen < tcpConf.packageMaxLength {
			Send(reg.Client, request.Flag | 4, request.Fd, string(rs))

		} else {
			for i := 0; i < dataLen; i += tcpConf.packageMaxLength {

				chunkLength := Min(tcpConf.packageMaxLength, dataLen - i)
				chunk := Substr(string(rs), i, chunkLength)

				flag := request.Flag
				if dataLen - i == chunkLength {
					flag |= 4
				}
				Send(reg.Client, flag, request.Fd, chunk)
			}
		}
	})

	reg.Client.Connect()
}

/**
	获取key
 */
func (reg *SRegister) getKey() string {
	return reg.ServiceData["pageServer"]["key"]
}

/**
	获取host
 */
func (reg *SRegister) getHost() string {
	return reg.ServiceData["pageServer"]["host"]
}

/**
	log4go对象设置
 */
func NewLog4go() *log4go.Logger {

	log4 := make(log4go.Logger)
	cw := log4go.NewConsoleLogWriter()

	// 非debug模式
	if isDebug == false {
		cw.SetFormat("[%T %D] [%L] %M")
	}
	log4.AddFilter("stdout", log4go.DEBUG, cw)
	log4.AddFilter("file", log4go.ERROR, log4go.NewFileLogWriter(logName, false))
	return &log4
}

/**
	设置debug状态
 */
func SetDebug(status bool) {
	isDebug = status
}

/**
	设置log日志文件路径
 */
func SetLogName(name string) {
	logName = name
}