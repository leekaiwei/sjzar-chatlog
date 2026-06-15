package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sjzar/chatlog/internal/chatlog/ctx"
	"github.com/sjzar/chatlog/internal/chatlog/database"
	"github.com/sjzar/chatlog/internal/chatlog/mcp"
	"github.com/sjzar/chatlog/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const (
	DefalutHTTPAddr = "127.0.0.1:5030"
)

type Service struct {
	ctx *ctx.Context
	db  *database.Service
	mcp *mcp.Service

	router *gin.Engine
	server *http.Server
}

func NewService(ctx *ctx.Context, db *database.Service, mcp *mcp.Service) *Service {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Handle error from SetTrustedProxies
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Err(err).Msg("Failed to set trusted proxies")
	}

	// Middleware
	router.Use(
		errors.RecoveryMiddleware(),
		errors.ErrorHandlerMiddleware(),
		gin.LoggerWithWriter(log.Logger),
	)

	s := &Service{
		ctx:    ctx,
		db:     db,
		mcp:    mcp,
		router: router,
	}

	router.Use(s.validateHostHeader())
	s.initRouter()
	return s
}

func (s *Service) validateHostHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, err := hostHeaderAllowed(c.Request.Host, s.ctx.HTTPAddr)
		if err != nil || !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid host"})
			return
		}
		c.Next()
	}
}

func hostHeaderAllowed(requestHost, serverAddr string) (bool, error) {
	if serverAddr == "" {
		serverAddr = DefalutHTTPAddr
	}
	_, serverPort, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return false, err
	}
	host, port, err := net.SplitHostPort(requestHost)
	if err != nil {
		return false, err
	}
	if port != serverPort {
		return false, nil
	}
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "localhost" {
		return true, nil
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback(), nil
}

func (s *Service) Start() error {

	if s.ctx.HTTPAddr == "" {
		s.ctx.HTTPAddr = DefalutHTTPAddr
	}

	if err := validateLoopbackAddr(s.ctx.HTTPAddr); err != nil {
		return err
	}

	s.server = &http.Server{
		Addr:    s.ctx.HTTPAddr,
		Handler: s.router,
	}

	go func() {
		// Handle error from Run
		if err := s.server.ListenAndServe(); err != nil {
			log.Err(err).Msg("Failed to start HTTP server")
		}
	}()

	log.Info().Msg("Starting HTTP server on " + s.ctx.HTTPAddr)

	return nil
}

func (s *Service) ListenAndServe() error {

	if s.ctx.HTTPAddr == "" {
		s.ctx.HTTPAddr = DefalutHTTPAddr
	}

	if err := validateLoopbackAddr(s.ctx.HTTPAddr); err != nil {
		return err
	}

	s.server = &http.Server{
		Addr:    s.ctx.HTTPAddr,
		Handler: s.router,
	}

	log.Info().Msg("Starting HTTP server on " + s.ctx.HTTPAddr)
	return s.server.ListenAndServe()
}

func validateLoopbackAddr(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid HTTP address %q: %w", addr, err)
	}
	if host == "" {
		return fmt.Errorf("invalid HTTP address %q: host is required", addr)
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if !ip.IsLoopback() {
			return fmt.Errorf("refusing to bind HTTP server to non-loopback host %q", host)
		}
		return nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("invalid HTTP address %q: %w", addr, err)
	}
	for _, resolved := range ips {
		if !resolved.IsLoopback() {
			return fmt.Errorf("refusing to bind HTTP server to non-loopback host %q", host)
		}
	}
	return nil
}

func (s *Service) Stop() error {

	if s.server == nil {
		return nil
	}

	// 使用超时上下文优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Debug().Err(err).Msg("Failed to shutdown HTTP server")
		return nil
	}

	log.Info().Msg("HTTP server stopped")
	return nil
}

func (s *Service) GetRouter() *gin.Engine {
	return s.router
}
