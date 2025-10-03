package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"

	"github.com/jantytgat/go-kit/outfit"
)

func NewHelloChanHandler(name string, maxWorkers int, idleTimeout time.Duration) *HelloChanHandler {
	return &HelloChanHandler{
		name:        name,
		maxWorkers:  maxWorkers,
		chMsg:       make(chan *nats.Msg, maxWorkers),
		idleTimeout: idleTimeout,
	}
}

type HelloChanHandler struct {
	name            string
	chMsg           chan *nats.Msg
	maxWorkers      int
	idleTimeout     time.Duration
	workerCtx       context.Context
	workerCtxCancel context.CancelFunc

	mux sync.Mutex
}

func (h *HelloChanHandler) getSubject(prefix string) string {
	if prefix != "" {
		return strings.Join([]string{prefix, h.name}, ".")
	}
	return h.name
}

func (h *HelloChanHandler) Subject(prefix string) string {
	return h.getSubject(prefix)
}

func (h *HelloChanHandler) Handler() chan *nats.Msg {
	h.mux.Lock()
	if h.chMsg == nil {
		h.chMsg = make(chan *nats.Msg, h.maxWorkers)
	}
	defer h.mux.Unlock()

	return h.chMsg
}

func (h *HelloChanHandler) MaxWorkers() int {
	return h.maxWorkers
}

func (h *HelloChanHandler) Handle(ctx context.Context, chMsg chan *nats.Msg) {
	fmt.Println("Handler started:", ctx.Value("id"))
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Handler done:", ctx.Value("id"))
			return
		case msg := <-chMsg:
			fmt.Println(string(msg.Data))
			// time.Sleep(200 * time.Millisecond)
		}
	}
}
func (h *HelloChanHandler) Start(ctx context.Context) {
	h.workerCtx, h.workerCtxCancel = context.WithCancel(ctx)
	for i := 0; i < h.maxWorkers; i++ {
		workerCtx := context.WithValue(ctx, "id", i+1)
		go h.Handle(workerCtx, h.chMsg)
	}
}

func (h *HelloChanHandler) Shutdown() {
	h.workerCtxCancel()
}

func main() {
	var err error
	var ns *server.Server

	var opts = &server.Options{
		ConfigFile:                 "",
		ServerName:                 "outfit-server",
		Host:                       "",
		Port:                       0,
		DontListen:                 true,
		ClientAdvertise:            "",
		Trace:                      true,
		Debug:                      true,
		TraceVerbose:               false,
		TraceHeaders:               false,
		NoLog:                      false,
		NoSigs:                     false,
		NoSublistCache:             false,
		NoHeaderSupport:            false,
		DisableShortFirstPing:      false,
		Logtime:                    false,
		LogtimeUTC:                 false,
		MaxConn:                    0,
		MaxSubs:                    0,
		MaxSubTokens:               0,
		Nkeys:                      nil,
		Users:                      nil,
		Accounts:                   nil,
		NoAuthUser:                 "",
		DefaultSentinel:            "",
		SystemAccount:              "",
		NoSystemAccount:            false,
		Username:                   "",
		Password:                   "",
		ProxyRequired:              false,
		Authorization:              "",
		AuthCallout:                nil,
		PingInterval:               0,
		MaxPingsOut:                0,
		HTTPHost:                   "",
		HTTPPort:                   0,
		HTTPBasePath:               "",
		HTTPSPort:                  0,
		AuthTimeout:                0,
		MaxControlLine:             0,
		MaxPayload:                 0,
		MaxPending:                 0,
		NoFastProducerStall:        false,
		Cluster:                    server.ClusterOpts{},
		Gateway:                    server.GatewayOpts{},
		LeafNode:                   server.LeafNodeOpts{},
		JetStream:                  false,
		NoJetStreamStrict:          false,
		JetStreamMaxMemory:         0,
		JetStreamMaxStore:          0,
		JetStreamDomain:            "",
		JetStreamExtHint:           "",
		JetStreamKey:               "",
		JetStreamOldKey:            "",
		JetStreamCipher:            0,
		JetStreamUniqueTag:         "",
		JetStreamLimits:            server.JSLimitOpts{},
		JetStreamTpm:               server.JSTpmOpts{},
		JetStreamMaxCatchup:        0,
		JetStreamRequestQueueLimit: 0,
		StreamMaxBufferedMsgs:      0,
		StreamMaxBufferedSize:      0,
		StoreDir:                   "",
		SyncInterval:               0,
		SyncAlways:                 false,
		JsAccDefaultDomain:         nil,
		Websocket:                  server.WebsocketOpts{},
		MQTT:                       server.MQTTOpts{},
		ProfPort:                   0,
		ProfBlockRate:              0,
		PidFile:                    "",
		PortsFileDir:               "",
		LogFile:                    "outfit-server.log",
		LogSizeLimit:               0,
		LogMaxFiles:                0,
		Syslog:                     false,
		RemoteSyslog:               "",
		Routes:                     nil,
		RoutesStr:                  "",
		TLSTimeout:                 0,
		TLS:                        false,
		TLSVerify:                  false,
		TLSMap:                     false,
		TLSCert:                    "",
		TLSKey:                     "",
		TLSCaCert:                  "",
		TLSConfig:                  nil,
		TLSPinnedCerts:             nil,
		TLSRateLimit:               0,
		TLSHandshakeFirst:          false,
		TLSHandshakeFirstFallback:  0,
		AllowNonTLS:                false,
		WriteDeadline:              0,
		MaxClosedClients:           0,
		LameDuckDuration:           0,
		LameDuckGracePeriod:        0,
		MaxTracedMsgLen:            0,
		TrustedKeys:                nil,
		TrustedOperators:           nil,
		AccountResolver:            nil,
		AccountResolverTLSConfig:   nil,
		AlwaysEnableNonce:          false,
		CustomClientAuthentication: nil,
		CustomRouterAuthentication: nil,
		CheckConfig:                false,
		DisableJetStreamBanner:     false,
		ConnectErrorReports:        0,
		ReconnectErrorReports:      0,
		Tags:                       nil,
		Metadata:                   nil,
		OCSPConfig:                 nil,
		Proxies:                    nil,
		OCSPCacheConfig:            nil,
	}
	if ns, err = server.NewServer(opts); err != nil {
		log.Fatal(err)
	}
	ns.ConfigureLogger()
	ns.Start()
	fmt.Printf("nats-server started: %s:%s\n", ns.ID(), ns.Name())
	defer ns.Shutdown()

	var ncOpts = nats.InProcessServer(ns)
	var nc *nats.Conn
	if nc, err = nats.Connect("", ncOpts); err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	fmt.Println("Connected to NATS server", nc.ConnectedServerName(), nc.ConnectedUrl(), nc.ConnectedServerId())

	handler := NewHelloChanHandler("hello", 10, 0)
	var module *outfit.Module
	if module, err = outfit.NewModule("module", nil); err != nil {
		log.Fatal(err)
	}

	if err = module.AddNatsChanHandler("", handler); err != nil {
		log.Fatal(err)
	}

	var subscriptions []*nats.Subscription
	if subscriptions, err = module.SubscribeAll("component", nc); err != nil {
		log.Fatal(err)
	}

	for _, sub := range subscriptions {
		fmt.Println("Subscribing", sub.Subject)
	}

	ctx, cancel := context.WithCancel(context.Background())
	handler.Start(ctx)

	fmt.Println("Sending messages")
	for i := 0; i < 100; i++ {
		if err = nc.Publish("component.module.hello", []byte(fmt.Sprintf("hello %d", i))); err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(1 * time.Second)
	fmt.Println("Client Stats", nc.Stats().InMsgs, nc.Stats().OutMsgs)
	cancel()
	time.Sleep(1 * time.Second)
}
