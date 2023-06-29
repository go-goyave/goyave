package goyave

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm/utils/tests"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/database"
	"goyave.dev/goyave/v4/util/fsutil"
)

type DummyService struct {
	AppName string
}

func (s *DummyService) Init(server *Server) {
	s.AppName = server.Config().GetString("app.name")
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
			fsutil.Delete("config.json")
		})

		s, err := New()
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
		assert.Equal(t, 20*time.Second, s.server.IdleTimeout)
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
			fsutil.Delete("config.json")
		})

		s, err := New()
		if assert.Error(t, err) {
			goyaveErr, ok := err.(*Error)
			if assert.True(t, ok) {
				assert.Equal(t, ExitInvalidConfig, goyaveErr.ExitCode)
			}
		}
		assert.Nil(t, s)
	})

	t.Run("NewWithConfig", func(t *testing.T) {
		database.RegisterDialect("sqlite3_server_test", "file:{name}?{options}", sqlite.Open)
		cfg := config.LoadDefault()
		cfg.Set("app.name", "test_with_config")
		cfg.Set("database.connection", "sqlite3_server_test")
		cfg.Set("database.name", "sqlite3_server_test.db")
		cfg.Set("database.options", "mode=memory")

		server, err := NewWithConfig(cfg)
		if !assert.NoError(t, err) {
			return
		}
		defer func() {
			assert.NoError(t, server.CloseDB())
		}()

		assert.Equal(t, "test_with_config", server.Config().GetString("app.name"))
		assert.NotNil(t, server.DB())

		assert.Nil(t, server.CloseDB())
	})

	t.Run("NewWithConfig_db_error", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("database.connection", "not_a_driver")

		server, err := NewWithConfig(cfg)
		if !assert.Error(t, err) {
			return
		}
		assert.Nil(t, server)
	})

	t.Run("getAddress", func(t *testing.T) {
		t.Run("0.0.0.0", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Set("server.host", "0.0.0.0")
			assert.Equal(t, "http://127.0.0.1:8080", getAddress(cfg))
		})
		t.Run("hide_port", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Set("server.port", 80)
			assert.Equal(t, "http://127.0.0.1", getAddress(cfg))
		})
		t.Run("domain", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Set("server.domain", "example.org")
			assert.Equal(t, "http://example.org:8080", getAddress(cfg))
		})
	})

	t.Run("getProxyAddress", func(t *testing.T) {
		t.Run("full", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Set("server.proxy.host", "proxy.example.org")
			cfg.Set("server.proxy.protocol", "https")
			cfg.Set("server.proxy.port", 1234)
			cfg.Set("server.proxy.base", "/base")
			assert.Equal(t, "https://proxy.example.org:1234/base", getProxyAddress(cfg))
		})

		t.Run("hide_port", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Set("server.proxy.host", "proxy.example.org")
			cfg.Set("server.proxy.protocol", "https")
			cfg.Set("server.proxy.port", 443)
			cfg.Set("server.proxy.base", "/base")
			assert.Equal(t, "https://proxy.example.org/base", getProxyAddress(cfg))

			cfg = config.LoadDefault()
			cfg.Set("server.proxy.host", "proxy.example.org")
			cfg.Set("server.proxy.protocol", "http")
			cfg.Set("server.proxy.port", 80)
			cfg.Set("server.proxy.base", "/base")
			assert.Equal(t, "http://proxy.example.org/base", getProxyAddress(cfg))
		})
	})

	t.Run("Service", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("app.name", "test")
		server, err := NewWithConfig(cfg)
		if !assert.NoError(t, err) {
			return
		}

		service := &DummyService{}
		server.RegisterService(service)
		assert.Equal(t, "test", service.AppName)
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
		server, err := NewWithConfig(cfg)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, "127.0.0.1:8080", server.Host())
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
		server, err := NewWithConfig(config.LoadDefault())
		if !assert.NoError(t, err) {
			return
		}

		server.RegisterRoutes(func(s *Server, router *Router) {
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
		server, err := NewWithConfig(cfg)
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
		server, err := NewWithConfig(config.LoadDefault())
		if !assert.NoError(t, err) {
			return
		}

		assert.Nil(t, server.ReplaceDB(tests.DummyDialector{}))
		assert.NotNil(t, server.db)
	})

	t.Run("Start", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("server.port", 8888)
		server, err := NewWithConfig(cfg)
		if !assert.NoError(t, err) {
			return
		}

		startupHookExecuted := false
		shutdownHookExecuted := false
		wg := sync.WaitGroup{}
		wg.Add(2)

		server.RegisterStartupHook(func(s *Server) {
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

		server.RegisterShutdownHook(func(s *Server) {
			shutdownHookExecuted = true
			assert.False(t, server.IsReady())
		})

		server.RegisterRoutes(func(s *Server, router *Router) {
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

	t.Run("Start_already_running", func(t *testing.T) {
		server, err := NewWithConfig(config.LoadDefault())
		if !assert.NoError(t, err) {
			return
		}
		atomic.StoreUint32(&server.state, 2) // Simulate the server already running
		err = server.Start()
		if assert.Error(t, err) {
			assert.Equal(t, "Server is already running", err.Error())
			e, ok := err.(*Error)
			if assert.True(t, ok) {
				assert.Equal(t, ExitStateError, e.ExitCode)
			}
		}
	})

	t.Run("Start_stopped", func(t *testing.T) {
		server, err := NewWithConfig(config.LoadDefault())
		if !assert.NoError(t, err) {
			return
		}
		atomic.StoreUint32(&server.state, 3) // Simulate stopped server
		err = server.Start()
		if assert.Error(t, err) {
			assert.Equal(t, "Cannot restart a stopped server", err.Error())
			e, ok := err.(*Error)
			if assert.True(t, ok) {
				assert.Equal(t, ExitStateError, e.ExitCode)
			}
		}
	})

	t.Run("Stop_not_started", func(t *testing.T) {
		server, err := NewWithConfig(config.LoadDefault())
		if !assert.NoError(t, err) {
			return
		}
		server.Stop()
		// Nothing happens
	})

	t.Run("StartupHooks", func(t *testing.T) {
		server, err := NewWithConfig(config.LoadDefault())
		if !assert.NoError(t, err) {
			return
		}

		server.RegisterStartupHook(func(s *Server) {})

		assert.Len(t, server.startupHooks, 1)

		server.ClearStartupHooks()
		assert.Len(t, server.startupHooks, 0)
	})

	t.Run("ShutdownHooks", func(t *testing.T) {
		server, err := NewWithConfig(config.LoadDefault())
		if !assert.NoError(t, err) {
			return
		}

		server.RegisterShutdownHook(func(s *Server) {})

		assert.Len(t, server.shutdownHooks, 1)

		server.ClearShutdownHooks()
		assert.Len(t, server.shutdownHooks, 0)
	})

	t.Run("SignalHook", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("server.port", 8889)
		server, err := NewWithConfig(cfg)
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

		server.RegisterStartupHook(func(s *Server) {
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
}
