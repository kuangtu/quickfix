package quickfix

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

//发起连接，处理session的消息
//Initiator initiates connections and processes messages for all sessions.
type Initiator struct {
	app             Application
	settings        *Settings
	sessionSettings map[SessionID]*SessionSettings
	storeFactory    MessageStoreFactory
	logFactory      LogFactory
	globalLog       Log
	stopChan        chan interface{}
	wg              sync.WaitGroup
	sessions        map[SessionID]*session
	sessionFactory
}

//Start Initiator.
func (i *Initiator) Start() (err error) {
	i.stopChan = make(chan interface{})

	for sessionID, settings := range i.sessionSettings {
		fmt.Println("the session Id is:", sessionID)
		//TODO: move into session factory
		var tlsConfig *tls.Config
		if tlsConfig, err = loadTLSConfig(settings); err != nil {
			return
		}

		var dialer proxy.Dialer
		if dialer, err = loadDialerConfig(settings); err != nil {
			return
		}

		//根据配置的sessionID，启动groutine
		i.wg.Add(1)
		go func(sessID SessionID) {
			//针对每个session进行处理
			i.handleConnection(i.sessions[sessID], tlsConfig, dialer)
			i.wg.Done()
		}(sessionID)
	}

	return
}

//Stop Initiator.
func (i *Initiator) Stop() {
	select {
	case <-i.stopChan:
		//closed already
		return
	default:
	}
	//程序退出，关闭stopChan
	close(i.stopChan)
	i.wg.Wait()
}

//NewInitiator creates and initializes a new Initiator.
func NewInitiator(app Application, storeFactory MessageStoreFactory, appSettings *Settings, logFactory LogFactory) (*Initiator, error) {
	i := &Initiator{
		app:             app,
		storeFactory:    storeFactory,
		settings:        appSettings,
		sessionSettings: appSettings.SessionSettings(),
		logFactory:      logFactory,
		sessions:        make(map[SessionID]*session),
		sessionFactory:  sessionFactory{true},
	}

	var err error
	i.globalLog, err = logFactory.Create()
	if err != nil {
		return i, err
	}

	for sessionID, s := range i.sessionSettings {
		session, err := i.createSession(sessionID, storeFactory, s, logFactory, app)
		if err != nil {
			return nil, err
		}

		i.sessions[sessionID] = session
	}

	return i, nil
}

//waitForInSessionTime returns true if the session is in session, false if the handler should stop
func (i *Initiator) waitForInSessionTime(session *session) bool {
	inSessionTime := make(chan interface{})
	go func() {
		session.waitForInSessionTime()
		close(inSessionTime)
	}()

	select {
	case <-inSessionTime:
	case <-i.stopChan:
		return false
	}

	return true
}

//waitForReconnectInterval returns true if a reconnect should be re-attempted, false if handler should stop
func (i *Initiator) waitForReconnectInterval(reconnectInterval time.Duration) bool {
	select {
	case <-time.After(reconnectInterval):
	case <-i.stopChan:
		return false
	}

	return true
}

//每个session处理连接
func (i *Initiator) handleConnection(session *session, tlsConfig *tls.Config, dialer proxy.Dialer) {
	var wg sync.WaitGroup
	wg.Add(1)
	//session的run方法是处理消息用的
	go func() {
		session.run()
		wg.Done()
	}()

	defer func() {
		session.stop()
		wg.Wait()
	}()

	connectionAttempt := 0

	for {
		if !i.waitForInSessionTime(session) {
			return
		}

		var disconnected chan interface{}
		var msgIn chan fixIn
		var msgOut chan []byte

		//开始连接对端地址
		address := session.SocketConnectAddress[connectionAttempt%len(session.SocketConnectAddress)]
		session.log.OnEventf("Connecting to: %v", address)
		fmt.Printf("Connecting to: %v", address)

		netConn, err := dialer.Dial("tcp", address)
		//连接失败进行重连
		if err != nil {
			session.log.OnEventf("Failed to connect: %v", err)
			goto reconnect
		} else if tlsConfig != nil {
			// Unless InsecureSkipVerify is true, server name config is required for TLS
			// to verify the received certificate
			if !tlsConfig.InsecureSkipVerify && len(tlsConfig.ServerName) == 0 {
				serverName := address
				if c := strings.LastIndex(serverName, ":"); c > 0 {
					serverName = serverName[:c]
				}
				tlsConfig.ServerName = serverName
			}
			//TLS连接
			tlsConn := tls.Client(netConn, tlsConfig)
			if err = tlsConn.Handshake(); err != nil {
				session.log.OnEventf("Failed handshake: %v", err)
				goto reconnect
			}
			netConn = tlsConn
		}

		//创建消息接收管道、消息发送管道
		msgIn = make(chan fixIn)
		msgOut = make(chan []byte)
		fmt.Println("create msgIn msgOut channel")
		if err := session.connect(msgIn, msgOut); err != nil {
			session.log.OnEventf("Failed to initiate: %v", err)
			goto reconnect
		}
		fmt.Println("send logon msg finished.")
		//通过bufio的NewReader读取socket中的内容，并通过newParser解析
		//其中接收到的数据通过管道传输
		go readLoop(newParser(bufio.NewReader(netConn)), msgIn)
		disconnected = make(chan interface{})
		//goroutine 写入socket
		go func() {
			writeLoop(netConn, msgOut, session.log)
			if err := netConn.Close(); err != nil {
				session.log.OnEvent(err.Error())
			}
			//如果
			close(disconnected)
		}()

		//如果中断或者stopChan退出
		select {
		case <-disconnected:
		case <-i.stopChan:
			return
		}

	reconnect:
		connectionAttempt++
		fmt.Println("reconnecting..")
		session.log.OnEventf("Reconnecting in %v", session.ReconnectInterval)
		if !i.waitForReconnectInterval(session.ReconnectInterval) {
			return
		}
	}
}
