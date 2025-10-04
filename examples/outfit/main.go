package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"

	"github.com/jantytgat/go-kit/outfit"
	"github.com/jantytgat/go-kit/slogd"
)

func main() {
	var err error
	slogd.Init(slogd.LevelDebug, false)
	slogd.RegisterSink(slogd.HandlerText, slog.NewTextHandler(os.Stdout, slogd.HandlerOptions()), true)
	logger := slogd.Logger().With(slog.String("service", "outfit"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// NATS-SERVER
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
	defer ns.Shutdown()
	logger.LogAttrs(ctx, slogd.LevelInfo, "nats-server started", slog.String("server-name", ns.Name()), slog.String("server-id", ns.ID()))

	// NATS-CLIENT
	var ncOpts = []nats.Option{
		nats.InProcessServer(ns),
		nats.ErrorHandler(natsErrHandler),
	}
	var nc *nats.Conn
	if nc, err = nats.Connect("", ncOpts...); err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	logger.LogAttrs(ctx, slogd.LevelInfo, "connected to nats server", slog.String("server-name", nc.ConnectedServerName()), slog.String("server-id", nc.ConnectedServerId()))

	// NATS HANDLERS
	var helloHandler *outfit.HelloHandler
	helloHandler = outfit.NewHelloHandler(ctx, "hello", "", 10000, logger)

	var handler1 *outfit.Handler
	handler1 = outfit.NewHandler("module1", helloHandler, logger)

	var module1 *outfit.Module
	if module1, err = outfit.NewModule(ctx, "module1", logger); err != nil {
		log.Fatal(err)
	}
	if err = module1.AddHandler(handler1); err != nil {
		log.Fatal(err)
	}

	// var subs []*nats.Subscription
	if err = module1.Subscribe(nc); err != nil {
		log.Fatal(err)
	}
	module1.Start(ctx)
	defer module1.Shutdown()

	for i := 0; i < 800000; i++ {
		if err = nc.Publish("hello", []byte(fmt.Sprintf("%d", i+1))); err != nil {
			log.Fatal(err)
		}
	}
}

func natsErrHandler(nc *nats.Conn, sub *nats.Subscription, natsErr error) {
	if errors.Is(natsErr, nats.ErrSlowConsumer) {
		switch sub.Type() {
		case nats.SyncSubscription:
			pendingMsgs, _, err := sub.Pending()
			if err != nil {
				fmt.Printf("couldn't get pending messages: %v\n", err)
				return
			}
			fmt.Printf("Falling behind with %d pending messages on subject %q.\n",
				pendingMsgs, sub.Subject)
			// Log error, notify operations...
		default:
			fmt.Println("NATS error:", natsErr, sub.Subject)
		}
	}
	// check for other errors
}
