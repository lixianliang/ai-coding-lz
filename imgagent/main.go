package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"imgagent/bailian"
	"imgagent/pkg/logger"
	"imgagent/svr"
)

var (
	confFile = flag.String("f", "imgagent.json", "image agent config filename")
)

type Config struct {
	LogConf         logger.Config      `json:"log_conf"`
	BindHost        string             `json:"bind_host"`
	BailianConf     bailian.Config     `json:"bailian"`
	DocumentMgrConf svr.DocumentConfig `json:"document_mgr"`

	svr.Config
}

func main() {
	flag.Parse()

	b, err := os.ReadFile(*confFile)
	if err != nil {
		log.Fatalf("Failed to ReadFile, err: %v", err)
	}
	var conf Config
	err = json.Unmarshal(b, &conf)
	if err != nil {
		log.Fatalf("Failed to Unmarshal, err: %v", err)
	}
	log.Println("conf: ", conf)

	_, err = logger.New(conf.LogConf)
	if err != nil {
		log.Fatalf("Failed to new logger, err: %v", err)
	}
	var wc io.WriteCloser = os.Stdout
	if conf.LogConf.AccessFile != "" {
		af, err := os.OpenFile(conf.LogConf.AccessFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Fatalf("Failed to OpenFile, err: %v", err)
		}
		wc = af
	}
	defer wc.Close()

	// 创建百炼客户端
	bailianClient, err := bailian.NewClient(conf.BailianConf)
	if err != nil {
		log.Fatalf("Failed to new bailian client, err: %v", err)
	}

	// 将百炼配置和文档管理配置传递给 Service
	conf.Config.BailianConfig = conf.BailianConf
	conf.Config.DocumentConfig = conf.DocumentMgrConf

	svr, err := svr.New(conf.Config, bailianClient)
	if err != nil {
		log.Fatalf("Failed to new server, err: %v", err)
	}

	router := svr.RegisterRouter(wc)
	server := &http.Server{
		Addr:              conf.BindHost,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	zap.S().Info("Server is running at ", conf.BindHost)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Fatalf("Server failed, err: %s", err)
		}
	}()

	SetupGracefulShutdown(server)
}

func SetupGracefulShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待中断信号
	<-quit
	zap.S().Info("Shutting down server...")

	// 设置5秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zap.S().Fatalf("Server forced to shutdown, err: %v", err)
	}

	zap.S().Info("Server exited gracefully")
}
