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
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"embed"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm/utils/tests"
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

		s, err := New(Options{
			MaxHeaderBytes: 123,
			ConnState:      func(_ net.Conn, _ http.ConnState) {},
			BaseContext:    func(_ net.Listener) context.Context { return context.Background() },
			ConnContext:    func(ctx context.Context, _ net.Conn) context.Context { return ctx },
		})
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, "test", s.Config().GetString("app.name"))
		assert.Nil(t, s.db)
		assert.NotNil(t, s.router)

		assert.NotNil(t, s.Lang)
		assert.Equal(t, "en-US", s.Lang.Default)
		assert.ElementsMatch(t, []string{"en-US", "en-UK"}, s.Lang.GetAvailableLanguages()) // All available languages are loaded

		assert.Equal(t, "127.0.0.1:8080", s.server.Addr)
		assert.Equal(t, 10*time.Second, s.server.WriteTimeout)
		assert.Equal(t, 10*time.Second, s.server.ReadTimeout)
		assert.Equal(t, 10*time.Second, s.server.ReadHeaderTimeout)
		assert.Equal(t, 20*time.Second, s.server.IdleTimeout)
		assert.Equal(t, 123, s.server.MaxHeaderBytes)
		assert.NotNil(t, s.server.ConnState)
		assert.NotNil(t, s.server.ConnContext)
		assert.NotNil(t, s.baseContext)
		assert.NotNil(t, s.server.BaseContext)
		assert.Equal(t, "http://127.0.0.1:8080", s.BaseURL())
		assert.Equal(t, "http://127.0.0.1:8080", s.ProxyBaseURL())
		assert.Nil(t, s.CloseDB())
	})

	t.Run("New_invalid_config", func(t *testing.T) {
		// Create a test config file (with only the app name)
		if err := os.WriteFile("config.json", []byte(`{"invalid"}`), 0644); err != nil {
			panic(err)
		}
		t.Cleanup(func() {
			if err := os.Remove("config.json"); err != nil {
				panic(err)
			}
		})

		s, err := New(Options{})
		if assert.Error(t, err) {
			goyaveErr, ok := err.(*errors.Error)
			if assert.True(t, ok) {
				assert.Equal(t, "Config error: invalid character '}' after object key", goyaveErr.Error())
			}
		}
		assert.Nil(t, s)
	})

	t.Run("NewWithOptions", func(t *testing.T) {
		database.RegisterDialect("sqlite3_server_test", "file:{name}?{options}", sqlite.Open)
		cfg := config.LoadDefault()
		cfg.Set("app.name", "test_with_config")
		cfg.Set("database.connection", "sqlite3_server_test")
		cfg.Set("database.name", "sqlite3_server_test.db")
		cfg.Set("database.options", "mode=memory")

		logger := slog.New(slog.NewHandler(false, &bytes.Buffer{}))
		langEmbed, err := fsutil.NewEmbed(resources).Sub("resources/lang")
		assert.NoError(t, err)
		opts := Options{
			Config: cfg,
			Logger: logger,
			LangFS: langEmbed,
		}

		server, err := New(opts)
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, server.CloseDB())
		}()

		assert.Equal(t, "test_with_config", server.Config().GetString("app.name"))
		assert.Equal(t, logger, server.Logger)
		assert.ElementsMatch(t, []string{"en-US", "en-UK"}, server.Lang.GetAvailableLanguages())
		assert.Equal(t, "load US", server.Lang.Get("en-US", "test-load"))
		assert.Equal(t, "load UK", server.Lang.Get("en-UK", "test-load"))
		assert.NotNil(t, server.DB())

		assert.Nil(t, server.CloseDB())
	})

	t.Run("NewWithConfig_db_error", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("database.connection", "not_a_driver")

		server, err := New(Options{Config: cfg})
		if !assert.Error(t, err) {
			return
		}
		assert.Nil(t, server)
	})

	t.Run("getAddress", func(t *testing.T) {
		t.Run("0.0.0.0", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Set("server.host", "0.0.0.0")
			cfg.Set("server.port", 8080)
			server := &Server{config: cfg, port: 8080}
			assert.Equal(t, "http://127.0.0.1:8080", server.getAddress(cfg))
		})
		t.Run("hide_port", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Set("server.port", 80)
			server := &Server{config: cfg, port: 80}
			assert.Equal(t, "http://127.0.0.1", server.getAddress(cfg))
		})
		t.Run("domain", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Set("server.domain", "example.org")
			server := &Server{config: cfg, port: 1234}
			assert.Equal(t, "http://example.org:1234", server.getAddress(cfg))
		})
	})

	t.Run("getProxyAddress", func(t *testing.T) {
		t.Run("full", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Set("server.proxy.host", "proxy.example.org")
			cfg.Set("server.proxy.protocol", "https")
			cfg.Set("server.proxy.port", 1234)
			cfg.Set("server.proxy.base", "/base")
			server := &Server{config: cfg, port: 1234}
			assert.Equal(t, "https://proxy.example.org:1234/base", server.getProxyAddress(cfg))
		})

		t.Run("hide_port", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Set("server.proxy.host", "proxy.example.org")
			cfg.Set("server.proxy.protocol", "https")
			cfg.Set("server.proxy.port", 443)
			cfg.Set("server.proxy.base", "/base")
			server := &Server{config: cfg, port: 443}
			assert.Equal(t, "https://proxy.example.org/base", server.getProxyAddress(cfg))

			cfg = config.LoadDefault()
			cfg.Set("server.proxy.host", "proxy.example.org")
			cfg.Set("server.proxy.protocol", "http")
			cfg.Set("server.proxy.port", 80)
			cfg.Set("server.proxy.base", "/base")
			server = &Server{config: cfg, port: 80}
			assert.Equal(t, "http://proxy.example.org/base", server.getProxyAddress(cfg))
		})
	})

	t.Run("Service", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("app.name", "test")
		server, err := New(Options{Config: cfg})
		if !assert.NoError(t, err) {
			return
		}

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
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, "127.0.0.1:8080", server.Host())
		assert.Equal(t, 8080, server.Port())
		assert.Equal(t, "http://127.0.0.1:8080", server.BaseURL())
		assert.Equal(t, "http://127.0.0.1:8080", server.ProxyBaseURL())
		assert.False(t, server.IsReady())
		assert.Equal(t, cfg, server.Config())
		assert.NotNil(t, server.Router())

		// No DB
		assert.Panics(t, func() {
			server.DB()
		})
	})

	t.Run("RegisterRoutes", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		if !assert.NoError(t, err) {
			return
		}

		server.RegisterRoutes(func(_ *Server, router *Router) {
			router.Get("/", func(_ *Response, _ *Request) {}).Name("base")
		})
		assert.NotNil(t, server.router.GetRoute("base"))
	})

	t.Run("Transaction", func(t *testing.T) {
		database.RegisterDialect("sqlite3_server_transaction_test", "file:{name}?{options}", sqlite.Open)
		cfg := config.LoadDefault()
		cfg.Set("database.connection", "sqlite3_server_transaction_test")
		cfg.Set("database.name", "sqlite3_server_transaction_test.db")
		cfg.Set("database.options", "mode=memory")
		server, err := New(Options{Config: cfg})
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, server.CloseDB())
		}()

		ogDB := server.db

		rollback := server.Transaction()

		assert.NotNil(t, rollback)
		assert.NotEqual(t, server.db, ogDB)

		rollback()
		assert.Equal(t, ogDB, server.db)

		assert.Panics(t, func() {
			server.db = nil
			server.Transaction()
		})
	})

	t.Run("ReplaceDB", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		if !assert.NoError(t, err) {
			return
		}

		assert.Nil(t, server.ReplaceDB(tests.DummyDialector{}))
		assert.NotNil(t, server.db)
	})

	t.Run("Start", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("server.port", 8888)
		server, err := New(Options{Config: cfg})
		if !assert.NoError(t, err) {
			return
		}

		startupHookExecuted := false
		shutdownHookExecuted := false
		wg := sync.WaitGroup{}
		wg.Add(2)

		server.RegisterStartupHook(func(_ *Server) {
			// Should be executed when the server is ready
			startupHookExecuted = true

			assert.True(t, server.IsReady())

			res, err := http.Get("http://localhost:8888")
			if !assert.NoError(t, err) {
				return
			}
			respBody, err := io.ReadAll(res.Body)
			if !assert.NoError(t, err) {
				return
			}
			_ = res.Body.Close()
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
			assert.Nil(t, err)
			wg.Done()
		}()

		wg.Wait()
		assert.True(t, startupHookExecuted)
		assert.True(t, shutdownHookExecuted)
		assert.False(t, server.IsReady())
		assert.Equal(t, uint32(3), atomic.LoadUint32(&server.state))
	})

	t.Run("StartWithAutoPort", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("server.port", 0)
		server, err := New(Options{Config: cfg})
		if !assert.NoError(t, err) {
			return
		}

		startupHookExecuted := false
		wg := sync.WaitGroup{}
		wg.Add(2)

		server.RegisterStartupHook(func(s *Server) {
			// Should be executed when the server is ready
			startupHookExecuted = true

			assert.True(t, server.IsReady())
			assert.NotEqual(t, 0, s.Port())

			res, err := http.Get(s.BaseURL())
			if !assert.NoError(t, err) {
				return
			}
			respBody, err := io.ReadAll(res.Body)
			if !assert.NoError(t, err) {
				return
			}
			_ = res.Body.Close()
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
			assert.Nil(t, err)
			wg.Done()
		}()

		wg.Wait()
		assert.True(t, startupHookExecuted)
		assert.False(t, server.IsReady())
		assert.Equal(t, uint32(3), atomic.LoadUint32(&server.state))
	})

	t.Run("Start_already_running", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		if !assert.NoError(t, err) {
			return
		}
		atomic.StoreUint32(&server.state, 2) // Simulate the server already running
		err = server.Start()
		if assert.Error(t, err) {
			assert.Equal(t, "server is already running", err.Error())
			_, ok := err.(*errors.Error)
			assert.True(t, ok)
		}
	})

	t.Run("Start_stopped", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		if !assert.NoError(t, err) {
			return
		}
		atomic.StoreUint32(&server.state, 3) // Simulate stopped server
		err = server.Start()
		if assert.Error(t, err) {
			assert.Equal(t, "cannot restart a stopped server", err.Error())
			_, ok := err.(*errors.Error)
			assert.True(t, ok)
		}
	})

	t.Run("Stop_not_started", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		if !assert.NoError(t, err) {
			return
		}
		server.Stop()
		// Nothing happens
	})

	t.Run("StartupHooks", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		if !assert.NoError(t, err) {
			return
		}

		server.RegisterStartupHook(func(_ *Server) {})

		assert.Len(t, server.startupHooks, 1)

		server.ClearStartupHooks()
		assert.Len(t, server.startupHooks, 0)
	})

	t.Run("ShutdownHooks", func(t *testing.T) {
		server, err := New(Options{Config: config.LoadDefault()})
		if !assert.NoError(t, err) {
			return
		}

		server.RegisterShutdownHook(func(_ *Server) {})

		assert.Len(t, server.shutdownHooks, 1)

		server.ClearShutdownHooks()
		assert.Len(t, server.shutdownHooks, 0)
	})

	t.Run("SignalHook", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("server.port", 8889)
		server, err := New(Options{Config: cfg})
		if !assert.NoError(t, err) {
			return
		}
		server.RegisterSignalHook()

		proc, err := os.FindProcess(os.Getpid())
		if !assert.NoError(t, err) {
			return
		}
		wg := sync.WaitGroup{}
		wg.Add(2)

		server.RegisterStartupHook(func(_ *Server) {
			if runtime.GOOS == "windows" {
				fmt.Println("Testing on a windows machine. Cannot test proc signals")
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
			assert.Nil(t, err)
			wg.Done()
		}()

		wg.Wait()
		assert.False(t, server.IsReady())
	})

	t.Run("Context", func(t *testing.T) {
		type baseContextKey struct{}
		type connContextKey struct{}

		cfg := config.LoadDefault()
		cfg.Set("server.port", 0)
		server, err := New(Options{
			Config: cfg,
			BaseContext: func(_ net.Listener) context.Context {
				return context.WithValue(context.Background(), baseContextKey{}, "base-ctx-value")
			},
			ConnContext: func(ctx context.Context, _ net.Conn) context.Context {
				return context.WithValue(ctx, connContextKey{}, "conn-ctx-value")
			},
		})
		if !assert.NoError(t, err) {
			return
		}

		startupHookExecuted := true
		wg := sync.WaitGroup{}
		wg.Add(2)

		server.RegisterStartupHook(func(s *Server) {
			// Should be executed when the server is ready
			startupHookExecuted = true

			assert.True(t, server.IsReady())

			res, err := http.Get(s.BaseURL())
			if !assert.NoError(t, err) {
				return
			}
			respBody, err := io.ReadAll(res.Body)
			if !assert.NoError(t, err) {
				return
			}
			_ = res.Body.Close()
			assert.Equal(t, []byte(fmt.Sprintf("%v|%v", "base-ctx-value", "conn-ctx-value")), respBody)

			// Stop the server, goroutine should return
			server.Stop()
			wg.Done()
		})

		server.RegisterRoutes(func(_ *Server, router *Router) {
			router.Get("/", func(r *Response, req *Request) {
				ctx := req.Context()
				assert.Equal(t, server, ServerFromContext(ctx))
				r.String(http.StatusOK, fmt.Sprintf("%v|%v", ctx.Value(baseContextKey{}), ctx.Value(connContextKey{})))
			}).Name("base")
		})

		go func() {
			err := server.Start()
			assert.Nil(t, err)
			wg.Done()
		}()

		wg.Wait()
		assert.True(t, startupHookExecuted)
		assert.False(t, server.IsReady())
		assert.Equal(t, uint32(3), atomic.LoadUint32(&server.state))
	})

	t.Run("NilBaseContext", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("server.port", 0)
		server, err := New(Options{
			Config: cfg,
			BaseContext: func(_ net.Listener) context.Context {
				return nil
			},
		})
		if !assert.NoError(t, err) {
			return
		}

		assert.Panics(t, func() {
			_ = server.Start()
		})
	})

	t.Run("StartServerWithCanceledContext", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("server.port", 0)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		server, err := New(Options{
			Config: cfg,
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
		})
		if !assert.NoError(t, err) {
			return
		}

		err = server.Start()
		if assert.Error(t, err) {
			assert.Equal(t, "cannot start the server, context is canceled", err.Error())
		}
	})
}

func TestNoServerFromContext(t *testing.T) {
	assert.Nil(t, ServerFromContext(context.Background()))
}

func TestErrLogWriter(t *testing.T) {

	s, err := New(Options{Config: config.LoadDefault()})
	if !assert.NoError(t, err) {
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	s.Logger = slog.New(slog.NewHandler(false, buf))

	w := &errLogWriter{
		server: s,
	}

	message := "error message"
	n, err := w.Write([]byte(message))
	assert.NoError(t, err)
	assert.Equal(t, len(message), n)

	assert.Regexp(t, regexp.MustCompile(
		fmt.Sprintf(`{"time":"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{1,9}((\+\d{2}:\d{2})|Z)?","level":"ERROR","source":{"function":".+","file":".+","line":\d+},"msg":"%s"}\n`,
			regexp.QuoteMeta(message),
		)),
		buf.String(),
	)
}
