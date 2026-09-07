package goyave

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"embed"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/database"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil"
)

//go:embed resources
var resources embed.FS

type DummyService struct {
	AppName string
}

func (s *DummyService) Name() string {
	return "dummy"
}

func TestServer(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		// Create a test config file (with only the app name)
		data, err := json.Marshal(map[string]any{"app": map[string]any{"name": "test"}})
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile("config.json", data, 0644); err != nil {
			panic(err)
		}
		t.Cleanup(func() {
			if err := os.Remove("config.json"); err != nil {
				panic(err)
			}
		})

		http2Cfg := &http.HTTP2Config{}
		customListenConfig := &net.ListenConfig{
			KeepAlive: 1 * time.Minute,
		}
		s, err := New(Options{
			MaxHeaderBytes: 123,
			ConnState:      func(_ net.Conn, _ http.ConnState) {},
			BaseContext:    func(_ context.Context, _ net.Listener) context.Context { return t.Context() },
			ConnContext:    func(ctx context.Context, _ net.Conn) context.Context { return ctx },
			HTTP2:          http2Cfg,
			ListenConfig:   customListenConfig,
		})
		require.NoError(t, err)

		// TODO update test
		// assert.Equal(t, "test", s.Config().GetString("app.name"))
		assert.Nil(t, s.db)
		assert.NotNil(t, s.router)

		assert.True(t, s.debug)
		assert.NotNil(t, s.Lang)
		assert.Equal(t, "en-US", s.Lang.Default)
		assert.ElementsMatch(t, []string{"en-US", "en-UK"}, s.Lang.GetAvailableLanguages()) // All available languages are loaded

		assert.Equal(t, "[::1]:8080", s.server.Addr)
		assert.Equal(t, 10*time.Second, s.server.WriteTimeout)
		assert.Equal(t, 10*time.Second, s.server.ReadTimeout)
		assert.Equal(t, 10*time.Second, s.server.ReadHeaderTimeout)
		assert.Equal(t, 20*time.Second, s.server.IdleTimeout)
		assert.Equal(t, 123, s.server.MaxHeaderBytes)
		assert.NotNil(t, s.server.ConnState)
		assert.NotNil(t, s.server.ConnContext)
		assert.NotNil(t, s.baseContext)
		assert.NotNil(t, s.server.BaseContext)
		assert.Same(t, http2Cfg, s.server.HTTP2)
		assert.Same(t, customListenConfig, s.listenConfig)
		assert.Equal(t, "http://[::1]:8080", s.BaseURL())
		assert.Equal(t, "http://[::1]:8080", s.ProxyBaseURL())
		assert.NoError(t, s.CloseDB())
		assert.NotNil(t, s.logger)

		// Logger and Server added to context
		assert.Equal(t, s.logger, slog.FromContext(s.Context()))
		assert.Equal(t, s, ServerFromContext(s.Context()))

		t.Run("ipv6_host", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Server.Host = "::"
			s, err = New(Options{Config: cfg})
			require.NoError(t, err)
			assert.Equal(t, "[::]:8080", s.server.Addr)
		})
	})

	t.Run("New_invalid_config", func(t *testing.T) {
		// Create a test config file (with only the app name)
		path := "config.json"
		if err := os.WriteFile(path, []byte(`{"invalid"}`), 0644); err != nil {
			panic(err)
		}
		t.Cleanup(func() {
			if err := os.Remove(path); err != nil {
				panic(err)
			}
		})

		s, err := New(Options{})
		if assert.Error(t, err) {
			var goyaveErr *errors.Error
			if assert.ErrorAs(t, err, &goyaveErr) {
				assert.ErrorContains(t, goyaveErr, "failed to unmarshal config")
				assert.ErrorContains(t, goyaveErr, "invalid character '}'")
			}
		}
		assert.Nil(t, s)
	})

	t.Run("NewWithOptions", func(t *testing.T) {
		database.RegisterDialect("sqlite3_server_test", "file:{name}?{options}", sqlite.Open)
		cfg := config.LoadDefault()
		// TODO update DB tests
		// cfg.Set("app.name", "test_with_config")
		// cfg.Set("database.connection", "sqlite3_server_test")
		// cfg.Set("database.name", "sqlite3_server_test.db")
		// cfg.Set("database.options", "mode=memory")

		logger := slog.New(slog.NewHandler(false, &bytes.Buffer{}))
		langEmbed, err := fsutil.NewEmbed(resources).Sub("resources/lang")
		require.NoError(t, err)
		opts := Options{
			Config: cfg,
			Logger: logger,
			LangFS: langEmbed,
		}

		server, err := New(opts)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, server.CloseDB())
		}()

		assert.Equal(t, logger, server.Logger())
		assert.ElementsMatch(t, []string{"en-US", "en-UK"}, server.Lang.GetAvailableLanguages())
		assert.Equal(t, "load US", server.Lang.Get("en-US", "test-load"))
		assert.Equal(t, "load UK", server.Lang.Get("en-UK", "test-load"))
		// TODO fix DB test
		// assert.NotNil(t, server.DB())
		// assert.True(t, server.HasDB())

		assert.NoError(t, server.CloseDB())
	})

	t.Run("Host", func(t *testing.T) {
		t.Run("ipv4", func(t *testing.T) {
			server := &Server{host: "0.0.0.0", port: 80}
			assert.Equal(t, "0.0.0.0:80", server.Host())
		})
		t.Run("ipv4_loopback", func(t *testing.T) {
			server := &Server{host: "127.0.0.1", port: 80}
			assert.Equal(t, "127.0.0.1:80", server.Host())
		})
		t.Run("ipv6", func(t *testing.T) {
			server := &Server{host: "::", port: 80}
			assert.Equal(t, "[::]:80", server.Host())
		})
		t.Run("ipv6_loopback", func(t *testing.T) {
			server := &Server{host: "::1", port: 80}
			assert.Equal(t, "[::1]:80", server.Host())
		})
	})

	t.Run("getAddress", func(t *testing.T) {
		t.Run("0.0.0.0", func(t *testing.T) {
			cfg := config.Server{}.Default()
			cfg.Host = "0.0.0.0"
			cfg.Port = 8080
			server := &Server{config: &cfg, port: 8080}
			assert.Equal(t, "http://127.0.0.1:8080", server.getAddress())
		})
		t.Run("0.0.0.0_ipv6", func(t *testing.T) {
			cfg := config.Server{}.Default()
			cfg.Host = "::"
			cfg.Port = 8080
			server := &Server{config: &cfg, port: 8080}
			assert.Equal(t, "http://[::1]:8080", server.getAddress())
		})
		t.Run("hide_port", func(t *testing.T) {
			cfg := config.Server{}.Default()
			cfg.Port = 80
			server := &Server{config: &cfg, port: 80}
			assert.Equal(t, "http://[::1]", server.getAddress())
		})
		t.Run("domain", func(t *testing.T) {
			cfg := config.Server{}.Default()
			cfg.Domain = "example.org"
			server := &Server{config: &cfg, port: 1234}
			assert.Equal(t, "http://example.org:1234", server.getAddress())
		})
		t.Run("ipv6", func(t *testing.T) {
			cfg := config.Server{}.Default()
			cfg.Host = "::1"
			server := &Server{config: &cfg, port: 1234}
			assert.Equal(t, "http://[::1]:1234", server.getAddress())
		})
		t.Run("ipv6_hide_port", func(t *testing.T) {
			cfg := config.Server{}.Default()
			cfg.Host = "::1"
			cfg.Port = 80
			server := &Server{config: &cfg, port: 80}
			assert.Equal(t, "http://[::1]", server.getAddress())
		})
	})

	t.Run("getProxyAddress", func(t *testing.T) {
		t.Run("full", func(t *testing.T) {
			cfg := config.Server{}.Default()
			cfg.Proxy.Host = "proxy.example.org"
			cfg.Proxy.Protocol = "https"
			cfg.Proxy.Port = 1234
			cfg.Proxy.Base = "/base"
			server := &Server{config: &cfg, port: 1234}
			assert.Equal(t, "https://proxy.example.org:1234/base", server.getProxyAddress())
		})

		t.Run("hide_port", func(t *testing.T) {
			cfg := config.Server{}.Default()
			cfg.Proxy.Host = "proxy.example.org"
			cfg.Proxy.Protocol = "https"
			cfg.Proxy.Port = 443
			cfg.Proxy.Base = "/base"
			server := &Server{config: &cfg, port: 443}
			assert.Equal(t, "https://proxy.example.org/base", server.getProxyAddress())

			cfg = config.Server{}.Default()
			cfg.Proxy.Host = "proxy.example.org"
			cfg.Proxy.Protocol = "http"
			cfg.Proxy.Port = 80
			cfg.Proxy.Base = "/base"
			server = &Server{config: &cfg, port: 80}
			assert.Equal(t, "http://proxy.example.org/base", server.getProxyAddress())
		})

		t.Run("full_ipv4", func(t *testing.T) {
			cfg := config.Server{}.Default()
			cfg.Proxy.Host = "192.168.1.11"
			cfg.Proxy.Protocol = "http"
			cfg.Proxy.Port = 1234
			cfg.Proxy.Base = "/base"
			server := &Server{config: &cfg, port: 1234}
			assert.Equal(t, "http://192.168.1.11:1234/base", server.getProxyAddress())
		})

		t.Run("full_ipv6", func(t *testing.T) {
			cfg := config.Server{}.Default()
			cfg.Proxy.Host = "::ffff:c0a8:10b"
			cfg.Proxy.Protocol = "http"
			cfg.Proxy.Port = 1234
			cfg.Proxy.Base = "/base"
			server := &Server{config: &cfg, port: 1234}
			assert.Equal(t, "http://[::ffff:c0a8:10b]:1234/base", server.getProxyAddress())
		})

		t.Run("hide_port_ipv6", func(t *testing.T) {
			cfg := config.Server{}.Default()
			cfg.Proxy.Host = "::ffff:c0a8:10b"
			cfg.Proxy.Protocol = "http"
			cfg.Proxy.Port = 80
			cfg.Proxy.Base = "/base"
			server := &Server{config: &cfg, port: 80}
			assert.Equal(t, "http://[::ffff:c0a8:10b]/base", server.getProxyAddress())
		})
	})

	t.Run("Service", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.App.Name = "test"
		server, err := New(Options{Config: cfg})
		require.NoError(t, err)

		service := &DummyService{}
		server.RegisterService(service)
		assert.Equal(t, map[string]Service{"dummy": service}, server.services)
		assert.Equal(t, service, server.Service("dummy"))

		s, ok := server.LookupService("dummy")
		assert.Equal(t, service, s)
		assert.True(t, ok)

		s, ok = server.LookupService("not_a_service")
		assert.Nil(t, s)
		assert.False(t, ok)

		assert.Panics(t, func() {
			server.Service("not_a_service")
		})
	})

	t.Run("Accessors", func(t *testing.T) {
		cfg := config.LoadDefault()
		server, err := New(Options{Config: cfg})
		require.NoError(t, err)

		assert.Equal(t, "[::1]:8080", server.Host())
		assert.Equal(t, 8080, server.Port())
		assert.Equal(t, "http://[::1]:8080", server.BaseURL())
		assert.Equal(t, "http://[::1]:8080", server.ProxyBaseURL())
		assert.False(t, server.IsReady())
		assert.NotNil(t, server.Router())
		assert.False(t, server.HasDB())
		assert.Equal(t, server.ctx, server.Context())
		assert.Equal(t, server.logger, server.Logger())

		// No DB
		assert.Panics(t, func() {
			server.DB()
		})
	})

	t.Run("RegisterRoutes", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		require.NoError(t, err)

		server.RegisterRoutes(func(_ *Server, router *Router) {
			router.Get("/", func(_ *Response, _ *Request) {}).Name("base")
		})
		assert.NotNil(t, server.router.GetRoute("base"))

		t.Run("panic_if_called_twice", func(t *testing.T) {
			assert.PanicsWithError(t, "router's regex cache has already been cleared, did you call RegisterRoutes twice?", func() {
				server.RegisterRoutes(func(_ *Server, _ *Router) {})
			})
		})
	})

	t.Run("Start", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Server.Port = 8888
		server, err := New(Options{Config: cfg})
		require.NoError(t, err)

		startupHookExecuted := false
		shutdownHookExecuted := false
		wg := sync.WaitGroup{}
		wg.Add(2)

		server.RegisterStartupHook(func(_ *Server) {
			// Should be executed when the server is ready
			startupHookExecuted = true

			assert.True(t, server.IsReady())

			res, err := http.Get("http://localhost:8888")
			defer func() {
				assert.NoError(t, res.Body.Close())
			}()
			assert.NoError(t, err)
			respBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, []byte("hello world"), respBody)

			// Stop the server, goroutine should return
			server.Stop()
			wg.Done()
		})

		server.RegisterShutdownHook(func(_ *Server) {
			shutdownHookExecuted = true
			assert.False(t, server.IsReady())
		})

		server.RegisterRoutes(func(_ *Server, router *Router) {
			router.Get("/", func(r *Response, _ *Request) {
				r.String(http.StatusOK, "hello world")
			}).Name("base")
		})

		go func() {
			err := server.Start()
			assert.NoError(t, err)
			wg.Done()
		}()

		wg.Wait()
		assert.True(t, startupHookExecuted)
		assert.True(t, shutdownHookExecuted)
		assert.False(t, server.IsReady())
		assert.Equal(t, uint32(3), server.state.Load())
	})

	t.Run("StartWithAutoPort", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Server.Port = 0
		server, err := New(Options{Config: cfg})
		require.NoError(t, err)

		startupHookExecuted := false
		wg := sync.WaitGroup{}
		wg.Add(2)

		server.RegisterStartupHook(func(s *Server) {
			// Should be executed when the server is ready
			startupHookExecuted = true

			assert.True(t, server.IsReady())
			assert.NotEqual(t, 0, s.Port())

			res, err := http.Get(s.BaseURL())
			defer func() {
				assert.NoError(t, res.Body.Close())
			}()
			assert.NoError(t, err)
			respBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, []byte("hello world"), respBody)

			// Stop the server, goroutine should return
			server.Stop()
			wg.Done()
		})

		server.RegisterRoutes(func(_ *Server, router *Router) {
			router.Get("/", func(r *Response, _ *Request) {
				r.String(http.StatusOK, "hello world")
			}).Name("base")
		})

		go func() {
			err := server.Start()
			assert.NoError(t, err)
			wg.Done()
		}()

		wg.Wait()
		assert.True(t, startupHookExecuted)
		assert.False(t, server.IsReady())
		assert.Equal(t, uint32(3), server.state.Load())
	})

	t.Run("Start_already_running", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		require.NoError(t, err)
		server.state.Store(2) // Simulate the server already running
		err = server.Start()
		if assert.Error(t, err) {
			assert.Equal(t, "server was already started", err.Error())
			_, ok := err.(*errors.Error)
			assert.True(t, ok)
		}
	})

	t.Run("Start_stopped", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		require.NoError(t, err)
		server.state.Store(3) // Simulate stopped server
		err = server.Start()
		if assert.Error(t, err) {
			assert.Equal(t, "server was already started", err.Error())
			_, ok := err.(*errors.Error)
			assert.True(t, ok)
		}
	})

	t.Run("Stop_not_started", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		assert.NoError(t, err)
		server.Stop()
		// Nothing happens
	})

	t.Run("Stop_already_stopped", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		assert.NoError(t, err)
		server.state.Store(3)
		server.Stop()
		// Nothing happens
	})

	t.Run("Stop_twice", func(t *testing.T) {
		// This test is for rare but possible cases of concurrent calls of
		// Stop(). There should be no error (sigChannel: close of closed channel)
		server, err := New(Options{Config: config.LoadDefault()})
		require.NoError(t, err)
		server.sigChannel = make(chan os.Signal, 64)
		assert.NotPanics(t, func() {
			server.Stop()
			server.Stop()
			// Nothing happens
		})
	})

	t.Run("StartupHooks", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		require.NoError(t, err)

		server.RegisterStartupHook(func(_ *Server) {})

		assert.Len(t, server.startupHooks, 1)

		server.ClearStartupHooks()
		assert.Empty(t, server.startupHooks)
	})

	t.Run("ShutdownHooks", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		require.NoError(t, err)

		server.RegisterShutdownHook(func(_ *Server) {})

		assert.Len(t, server.shutdownHooks, 1)

		server.ClearShutdownHooks()
		assert.Empty(t, server.shutdownHooks)
	})

	t.Run("SignalHook", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Server.Port = 8889
		server, err := New(Options{Config: cfg})
		require.NoError(t, err)
		server.RegisterSignalHook()

		proc, err := os.FindProcess(os.Getpid())
		require.NoError(t, err)
		wg := sync.WaitGroup{}
		wg.Add(2)

		server.RegisterStartupHook(func(_ *Server) {
			if runtime.GOOS == "windows" {
				t.Logf("Testing on a windows machine. Cannot test proc signals")
				server.Stop()
			} else {
				time.Sleep(10 * time.Millisecond)
				if err := proc.Signal(syscall.SIGTERM); err != nil {
					assert.Fail(t, err.Error())
				}
			}
			wg.Done()
		})

		go func() {
			err := server.Start()
			assert.NoError(t, err)
			wg.Done()
		}()

		wg.Wait()
		assert.False(t, server.IsReady())
	})

	t.Run("Context", func(t *testing.T) {
		type rootContextKey struct{}
		type baseContextKey struct{}
		type connContextKey struct{}

		rootCtx := context.WithValue(t.Context(), rootContextKey{}, "root-ctx-value")

		cfg := config.LoadDefault()
		cfg.Server.Port = 0
		server, err := New(Options{
			Config:  cfg,
			Context: rootCtx,
			BaseContext: func(ctx context.Context, _ net.Listener) context.Context {
				return context.WithValue(ctx, baseContextKey{}, "base-ctx-value")
			},
			ConnContext: func(ctx context.Context, _ net.Conn) context.Context {
				return context.WithValue(ctx, connContextKey{}, "conn-ctx-value")
			},
		})
		require.NoError(t, err)

		startupHookExecuted := true
		wg := sync.WaitGroup{}
		wg.Add(2)

		server.RegisterStartupHook(func(s *Server) {
			// Should be executed when the server is ready
			startupHookExecuted = true

			assert.True(t, server.IsReady())

			res, err := http.Get(s.BaseURL())
			defer func() {
				assert.NoError(t, res.Body.Close())
			}()
			assert.NoError(t, err)
			respBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("%s|%s|%s", "root-ctx-value", "base-ctx-value", "conn-ctx-value"), string(respBody))

			// Stop the server, goroutine should return
			server.Stop()
			wg.Done()
		})

		server.RegisterRoutes(func(_ *Server, router *Router) {
			router.Get("/", func(r *Response, req *Request) {
				ctx := req.Context()
				assert.Equal(t, server, ServerFromContext(ctx))
				assert.Equal(t, server.Logger(), slog.FromContext(ctx))
				r.String(http.StatusOK, fmt.Sprintf("%v|%v|%v", ctx.Value(rootContextKey{}), ctx.Value(baseContextKey{}), ctx.Value(connContextKey{})))
			}).Name("base")
		})

		go func() {
			err := server.Start()
			assert.NoError(t, err)
			wg.Done()
		}()

		wg.Wait()
		assert.True(t, startupHookExecuted)
		assert.False(t, server.IsReady())
		assert.Equal(t, uint32(3), server.state.Load())
	})

	t.Run("NilBaseContext", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Server.Port = 0
		server, err := New(Options{
			Config: cfg,
			BaseContext: func(_ context.Context, _ net.Listener) context.Context {
				return nil
			},
		})
		require.NoError(t, err)

		assert.Panics(t, func() {
			_ = server.Start() // TODO only called on new connection now
		})
	})

	t.Run("StartServerWithCanceledContext", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Server.Port = 0
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		server, err := New(Options{
			Config:  cfg,
			Context: ctx,
		})
		require.NoError(t, err)

		err = server.Start()
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
		assert.ErrorContains(t, err, "cannot start the server, context is canceled")
	})

	t.Run("StartWithCustomListenConfig", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Server.Port = 0

		customListenConfig := &net.ListenConfig{
			KeepAlive: 1 * time.Minute,
		}

		server, err := New(Options{
			Config:       cfg,
			ListenConfig: customListenConfig,
		})
		require.NoError(t, err)

		wg := sync.WaitGroup{}
		wg.Add(2)

		server.RegisterStartupHook(func(s *Server) {
			assert.True(t, server.IsReady())
			assert.NotEqual(t, 0, s.Port())

			res, err := http.Get(s.BaseURL())
			defer func() {
				assert.NoError(t, res.Body.Close())
			}()
			assert.NoError(t, err)
			respBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, []byte("hello world"), respBody)

			server.Stop()
			wg.Done()
		})

		server.RegisterRoutes(func(_ *Server, router *Router) {
			router.Get("/", func(r *Response, _ *Request) {
				r.String(http.StatusOK, "hello world")
			}).Name("base")
		})

		go func() {
			err := server.Start()
			assert.NoError(t, err)
			wg.Done()
		}()

		wg.Wait()
		assert.False(t, server.IsReady())
		assert.Equal(t, uint32(3), server.state.Load())
	})

	t.Run("StartWithCustomListenConfigControlError", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Server.Port = 0

		// Create a custom ListenConfig with a Control function that always returns an error
		// to ensure that the custom config is correctly used if provided.
		// The StartWithCustomListenConfig test only checks that the server still works and is able
		// to process requests with the custom config.
		expectedErr := fmt.Errorf("test control error")
		customListenConfig := &net.ListenConfig{
			Control: func(_, _ string, _ syscall.RawConn) error {
				return expectedErr
			},
		}

		server, err := New(Options{
			Config:       cfg,
			ListenConfig: customListenConfig,
		})
		require.NoError(t, err)

		// Attempt to start the server - it should fail with the error from Control
		err = server.Start()
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestNoServerFromContext(t *testing.T) {
	assert.Nil(t, ServerFromContext(context.Background()))
}

func TestErrLogWriter(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	logger := slog.New(slog.NewHandler(false, buf))
	s, err := New(Options{Config: config.LoadDefault(), Logger: logger})
	require.NoError(t, err)

	w := &errLogWriter{
		server: s,
	}

	message := "error message"
	n, err := w.Write([]byte(message))
	require.NoError(t, err)
	assert.Equal(t, len(message), n)

	assert.Regexp(t,
		fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s"}\n`,
			regexp.QuoteMeta(message),
		),
		buf.String(),
	)
}
