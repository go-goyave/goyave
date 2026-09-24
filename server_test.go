package goyave

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"embed"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/semconv/v1.43.0/httpconv"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"gorm.io/driver/sqlite"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/database"
	"goyave.dev/goyave/v5/internal/otel"
	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/errwrap"
	"goyave.dev/goyave/v5/util/fsutil"
)

//go:embed resources
var resources embed.FS

func TestServer(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		http2Cfg := &http.HTTP2Config{}
		customListenConfig := &net.ListenConfig{
			KeepAlive: 1 * time.Minute,
		}
		cfg := config.LoadDefault()
		s, err := New(cfg, Options{
			Context:               t.Context(),
			MaxHeaderBytes:        123,
			MaxHeaderValueCount:   123,
			DisableClientPriority: true,
			ConnState:             func(_ net.Conn, _ http.ConnState) {},
			BaseContext:           func(_ context.Context, _ net.Listener) context.Context { return t.Context() },
			ConnContext:           func(ctx context.Context, _ net.Conn) context.Context { return ctx },
			HTTP2:                 http2Cfg,
			ListenConfig:          customListenConfig,
		})
		require.NoError(t, err)

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
		assert.Equal(t, 123, s.server.MaxHeaderValueCount)
		assert.True(t, s.server.DisableClientPriority)
		assert.NotNil(t, s.server.ConnState)
		assert.NotNil(t, s.server.ConnContext)
		assert.NotNil(t, s.baseContext)
		assert.NotNil(t, s.server.BaseContext)
		assert.Same(t, http2Cfg, s.server.HTTP2)
		assert.Same(t, customListenConfig, s.listenConfig)
		assert.Equal(t, "http://[::1]:8080", s.BaseURL())
		assert.Equal(t, "http://[::1]:8080", s.ProxyBaseURL())
		assert.NotNil(t, s.logger)

		// Logger and Server added to context
		assert.Equal(t, s.logger, slog.FromContext(s.Context()))
		assert.Same(t, s, ServerFromContext(s.Context()))

		t.Run("ipv6_host", func(t *testing.T) {
			cfg := config.LoadDefault()
			cfg.Server.Host = "::"
			s, err = New(cfg, Options{})
			require.NoError(t, err)
			assert.Equal(t, "[::]:8080", s.server.Addr)
		})
	})

	t.Run("NewWithOptions", func(t *testing.T) {
		database.RegisterDialect("sqlite3_server_test", "file:{name}?{options}", sqlite.Open)
		cfg := config.LoadDefault()

		logger := slog.New(slog.NewHandler(false, &bytes.Buffer{}))
		langEmbed, err := fsutil.NewEmbed(resources).Sub("resources/lang")
		require.NoError(t, err)
		opts := Options{
			Logger: logger,
			LangFS: langEmbed,
			OpenTelemetry: OpenTelemetryOptions{
				TracerProvider: noop.NewTracerProvider(),
			},
		}

		server, err := New(cfg, opts)
		require.NoError(t, err)

		assert.Equal(t, logger, server.Logger())
		assert.ElementsMatch(t, []string{"en-US", "en-UK"}, server.Lang.GetAvailableLanguages())
		assert.Equal(t, "load US", server.Lang.Get("en-US", "test-load"))
		assert.Equal(t, "load UK", server.Lang.Get("en-UK", "test-load"))
		assert.NotNil(t, server.otelTracer)
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

	t.Run("Accessors", func(t *testing.T) {
		cfg := config.LoadDefault()
		server, err := New(cfg, Options{})
		require.NoError(t, err)

		assert.Equal(t, "[::1]:8080", server.Host())
		assert.Equal(t, 8080, server.Port())
		assert.Equal(t, "http://[::1]:8080", server.BaseURL())
		assert.Equal(t, "http://[::1]:8080", server.ProxyBaseURL())
		assert.False(t, server.IsReady())
		assert.NotNil(t, server.Router())
		assert.Equal(t, server.ctx, server.Context())
		assert.Equal(t, server.logger, server.Logger())
	})

	t.Run("Start", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Server.Port = 8888
		server, err := New(cfg, Options{})
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

		server.Router().Get("/", func(r *Response, _ *Request) {
			r.String(http.StatusOK, "hello world")
		}).Name("base")

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
		server, err := New(cfg, Options{})
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

		server.Router().Get("/", func(r *Response, _ *Request) {
			r.String(http.StatusOK, "hello world")
		}).Name("base")

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
		server, err := New(config.LoadDefault(), Options{})
		require.NoError(t, err)
		server.state.Store(2) // Simulate the server already running
		err = server.Start()
		if assert.Error(t, err) {
			assert.Equal(t, "server was already started", err.Error())
			_, ok := err.(*errwrap.Error)
			assert.True(t, ok)
		}
	})

	t.Run("Start_stopped", func(t *testing.T) {
		server, err := New(config.LoadDefault(), Options{})
		require.NoError(t, err)
		server.state.Store(3) // Simulate stopped server
		err = server.Start()
		if assert.Error(t, err) {
			assert.Equal(t, "server was already started", err.Error())
			_, ok := err.(*errwrap.Error)
			assert.True(t, ok)
		}
	})

	t.Run("Stop_not_started", func(t *testing.T) {
		server, err := New(config.LoadDefault(), Options{})
		assert.NoError(t, err)
		server.Stop()
		// Nothing happens
	})

	t.Run("Stop_already_stopped", func(t *testing.T) {
		server, err := New(config.LoadDefault(), Options{})
		assert.NoError(t, err)
		server.state.Store(3)
		server.Stop()
		// Nothing happens
	})

	t.Run("Stop_twice", func(t *testing.T) {
		// This test is for rare but possible cases of concurrent calls of
		// Stop(). There should be no error (sigChannel: close of closed channel)
		server, err := New(config.LoadDefault(), Options{})
		require.NoError(t, err)
		server.sigChannel = make(chan os.Signal, 64)
		assert.NotPanics(t, func() {
			server.Stop()
			server.Stop()
			// Nothing happens
		})
	})

	t.Run("StartupHooks", func(t *testing.T) {
		server, err := New(config.LoadDefault(), Options{})
		require.NoError(t, err)

		server.RegisterStartupHook(func(_ *Server) {})

		assert.Len(t, server.startupHooks, 1)

		server.ClearStartupHooks()
		assert.Empty(t, server.startupHooks)
	})

	t.Run("ShutdownHooks", func(t *testing.T) {
		server, err := New(config.LoadDefault(), Options{})
		require.NoError(t, err)

		server.RegisterShutdownHook(func(_ *Server) {})

		assert.Len(t, server.shutdownHooks, 1)

		server.ClearShutdownHooks()
		assert.Empty(t, server.shutdownHooks)
	})

	t.Run("SignalHook", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Server.Port = 8889
		server, err := New(cfg, Options{})
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
		server, err := New(cfg, Options{
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

		server.Router().Get("/", func(r *Response, req *Request) {
			ctx := req.Context()
			assert.Equal(t, server, ServerFromContext(ctx))
			assert.Equal(t, server.Logger(), slog.FromContext(ctx))
			r.String(http.StatusOK, fmt.Sprintf("%v|%v|%v", ctx.Value(rootContextKey{}), ctx.Value(baseContextKey{}), ctx.Value(connContextKey{})))
		}).Name("base")

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
		server, err := New(cfg, Options{
			BaseContext: func(_ context.Context, _ net.Listener) context.Context {
				return nil
			},
		})
		require.NoError(t, err)

		assert.Panics(t, func() {
			// std http.Server calls BaseContext on start
			_ = server.Start()
		})
	})

	t.Run("StartServerWithCanceledContext", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Server.Port = 0
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		server, err := New(cfg, Options{
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

		server, err := New(cfg, Options{
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

		server.Router().Get("/", func(r *Response, _ *Request) {
			r.String(http.StatusOK, "hello world")
		}).Name("base")

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

		server, err := New(cfg, Options{
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
	s, err := New(config.LoadDefault(), Options{Logger: logger})
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

func TestOpenTelemetry(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			spanRecorder, metrics := prepareOpenTelemetryTest(t, "/uri/test", func(response *Response, request *Request) {
				// Drain the body to see if the request body size metric is reported
				_, _ = io.ReadAll(request.Body())

				span := trace.SpanFromContext(request.Context())
				span.SetAttributes(attribute.Bool("handler_reached", true))

				bag := baggage.FromContext(request.Context())
				assert.Equal(t, "alice", bag.Member("userId").Value())
				assert.Equal(t, "false", bag.Member("isProduction").Value())

				// Advance clock to check the request duration metric
				time.Sleep(time.Second*3 + time.Millisecond*200)

				response.String(http.StatusOK, "hello world")
			})

			spans := spanRecorder.Ended()
			require.Len(t, spans, 1)
			span := spans[0]
			assert.Equal(t, "GET /uri/{param}", span.Name())

			status := span.Status()
			assert.Equal(t, codes.Unset, status.Code)
			assert.Empty(t, status.Description)

			parent := span.Parent()
			assert.Equal(t, "0af7651916cd43dd8448eb211c80319c", parent.TraceID().String())
			assert.Equal(t, "b7ad6b7169203331", parent.SpanID().String())
			assert.Equal(t, "01", parent.TraceFlags().String())
			assert.Equal(t, "congo=t61rcWkgMzE", parent.TraceState().String())
			assert.True(t, parent.IsRemote())

			wantAttrs := []attribute.KeyValue{
				semconv.HTTPRequestMethodGet,
				semconv.HTTPRoute("/uri/{param}"),
				semconv.ServerAddress("example.com"),
				semconv.ClientAddress("192.0.2.1"),
				semconv.ClientPort(1234),
				semconv.URLFull("/uri/test"), // In a test environment, we don't have a full URL with proto and host.
				semconv.URLPath("/uri/test"),
				semconv.NetworkProtocolVersion("1.1"),
				semconv.NetworkPeerAddress("192.0.2.1"),
				semconv.NetworkPeerPort(1234),
				attribute.Bool("handler_reached", true),
				semconv.HTTPResponseStatusCode(http.StatusOK),
			}
			assert.Equal(t, wantAttrs, span.Attributes())
			events := span.Events()
			assert.Empty(t, events)

			scope := span.InstrumentationScope()
			assert.Equal(t, otel.OpenTelemetryTracerName, scope.Name)
			assert.Equal(t, otel.Version, scope.Version)

			if assert.Len(t, metrics.ScopeMetrics, 1) {
				sm := metrics.ScopeMetrics[0]
				assert.Equal(t, otel.OpenTelemetryMeterName, sm.Scope.Name)
				assert.Equal(t, otel.Version, sm.Scope.Version)

				if assert.Len(t, sm.Metrics, 3) {
					requestBodySizeMetric := sm.Metrics[0]
					responseBodySizeMetric := sm.Metrics[1]
					requestDurationMetric := sm.Metrics[2]

					assertIntMetric(t, requestBodySizeMetric, httpconv.ServerRequestBodySize{}.Name(), 5, http.StatusOK, nil, false)
					assertIntMetric(t, responseBodySizeMetric, httpconv.ServerResponseBodySize{}.Name(), 11, http.StatusOK, nil, false)
					assertDurationMetric(t, requestDurationMetric, httpconv.ServerRequestDuration{}.Name(), 3.2, http.StatusOK, nil, false)
				}
			}
		})
	})

	t.Run("error", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			spanRecorder, metrics := prepareOpenTelemetryTest(t, "/uri/test", func(response *Response, _ *Request) {
				time.Sleep(time.Second*3 + time.Millisecond*200)
				response.Error("test error")
			})

			spans := spanRecorder.Ended()
			require.Len(t, spans, 1)
			span := spans[0]
			assert.Equal(t, "GET /uri/{param}", span.Name())

			status := span.Status()
			assert.Equal(t, codes.Error, status.Code)
			assert.Equal(t, "test error", status.Description)

			parent := span.Parent()
			assert.Equal(t, "0af7651916cd43dd8448eb211c80319c", parent.TraceID().String())
			assert.Equal(t, "b7ad6b7169203331", parent.SpanID().String())
			assert.Equal(t, "01", parent.TraceFlags().String())
			assert.Equal(t, "congo=t61rcWkgMzE", parent.TraceState().String())
			assert.True(t, parent.IsRemote())

			wantAttrs := []attribute.KeyValue{
				semconv.HTTPRequestMethodGet,
				semconv.HTTPRoute("/uri/{param}"),
				semconv.ServerAddress("example.com"),
				semconv.ClientAddress("192.0.2.1"),
				semconv.ClientPort(1234),
				semconv.URLFull("/uri/test"), // In a test environment, we don't have a full URL with proto and host.
				semconv.URLPath("/uri/test"),
				semconv.NetworkProtocolVersion("1.1"),
				semconv.NetworkPeerAddress("192.0.2.1"),
				semconv.NetworkPeerPort(1234),
				semconv.HTTPResponseStatusCode(http.StatusInternalServerError),
			}
			assert.Equal(t, wantAttrs, span.Attributes())

			scope := span.InstrumentationScope()
			assert.Equal(t, otel.OpenTelemetryTracerName, scope.Name)
			assert.Equal(t, otel.Version, scope.Version)

			if assert.Len(t, metrics.ScopeMetrics, 1) {
				sm := metrics.ScopeMetrics[0]
				assert.Equal(t, otel.OpenTelemetryMeterName, sm.Scope.Name)
				assert.Equal(t, otel.Version, sm.Scope.Version)

				if assert.Len(t, sm.Metrics, 3) {
					requestBodySizeMetric := sm.Metrics[0]
					responseBodySizeMetric := sm.Metrics[1]
					requestDurationMetric := sm.Metrics[2]

					assertIntMetric(t, requestBodySizeMetric, httpconv.ServerRequestBodySize{}.Name(), 0, http.StatusInternalServerError, &errwrap.Error{}, false)
					assertIntMetric(t, responseBodySizeMetric, httpconv.ServerResponseBodySize{}.Name(), 22, http.StatusInternalServerError, &errwrap.Error{}, false)
					assertDurationMetric(t, requestDurationMetric, httpconv.ServerRequestDuration{}.Name(), 3.2, http.StatusInternalServerError, &errwrap.Error{}, false)
				}
			}
		})
	})

	t.Run("panic", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			spanRecorder, metrics := prepareOpenTelemetryTest(t, "/uri/test", func(_ *Response, _ *Request) {
				time.Sleep(time.Second*3 + time.Millisecond*200)
				panic("test error")
			})

			spans := spanRecorder.Ended()
			require.Len(t, spans, 1)
			span := spans[0]
			assert.Equal(t, "GET /uri/{param}", span.Name())

			status := span.Status()
			assert.Equal(t, codes.Error, status.Code)
			assert.Equal(t, "test error", status.Description)

			parent := span.Parent()
			assert.Equal(t, "0af7651916cd43dd8448eb211c80319c", parent.TraceID().String())
			assert.Equal(t, "b7ad6b7169203331", parent.SpanID().String())
			assert.Equal(t, "01", parent.TraceFlags().String())
			assert.Equal(t, "congo=t61rcWkgMzE", parent.TraceState().String())
			assert.True(t, parent.IsRemote())

			wantAttrs := []attribute.KeyValue{
				semconv.HTTPRequestMethodGet,
				semconv.HTTPRoute("/uri/{param}"),
				semconv.ServerAddress("example.com"),
				semconv.ClientAddress("192.0.2.1"),
				semconv.ClientPort(1234),
				semconv.URLFull("/uri/test"), // In a test environment, we don't have a full URL with proto and host.
				semconv.URLPath("/uri/test"),
				semconv.NetworkProtocolVersion("1.1"),
				semconv.NetworkPeerAddress("192.0.2.1"),
				semconv.NetworkPeerPort(1234),
				semconv.HTTPResponseStatusCode(http.StatusInternalServerError),
			}
			assert.Equal(t, wantAttrs, span.Attributes())

			scope := span.InstrumentationScope()
			assert.Equal(t, otel.OpenTelemetryTracerName, scope.Name)
			assert.Equal(t, otel.Version, scope.Version)

			if assert.Len(t, metrics.ScopeMetrics, 1) {
				sm := metrics.ScopeMetrics[0]
				assert.Equal(t, otel.OpenTelemetryMeterName, sm.Scope.Name)
				assert.Equal(t, otel.Version, sm.Scope.Version)

				if assert.Len(t, sm.Metrics, 3) {
					requestBodySizeMetric := sm.Metrics[0]
					responseBodySizeMetric := sm.Metrics[1]
					requestDurationMetric := sm.Metrics[2]

					assertIntMetric(t, requestBodySizeMetric, httpconv.ServerRequestBodySize{}.Name(), 0, http.StatusInternalServerError, &errwrap.Error{}, false)
					assertIntMetric(t, responseBodySizeMetric, httpconv.ServerResponseBodySize{}.Name(), 22, http.StatusInternalServerError, &errwrap.Error{}, false)
					assertDurationMetric(t, requestDurationMetric, httpconv.ServerRequestDuration{}.Name(), 3.2, http.StatusInternalServerError, &errwrap.Error{}, false)
				}
			}
		})
	})

	t.Run("protocol_redirect", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			spanRecorder, metrics := prepareOpenTelemetryTest(t, "https://example.com:8080/uri/test", func(_ *Response, _ *Request) {
				// We shouldn't reach this handler, the expected request duration is therefore 0
				time.Sleep(time.Second*3 + time.Millisecond*200)
				panic("test error")
			})

			spans := spanRecorder.Ended()
			require.Len(t, spans, 1)
			span := spans[0]
			assert.Equal(t, "GET", span.Name())

			status := span.Status()
			assert.Equal(t, codes.Unset, status.Code)
			assert.Empty(t, status.Description)

			parent := span.Parent()
			assert.Equal(t, "0af7651916cd43dd8448eb211c80319c", parent.TraceID().String())
			assert.Equal(t, "b7ad6b7169203331", parent.SpanID().String())
			assert.Equal(t, "01", parent.TraceFlags().String())
			assert.Equal(t, "congo=t61rcWkgMzE", parent.TraceState().String())
			assert.True(t, parent.IsRemote())

			wantAttrs := []attribute.KeyValue{
				semconv.HTTPRequestMethodGet,
				semconv.URLScheme("https"),
				semconv.ServerAddress("example.com"),
				semconv.ServerPort(8080),
				semconv.ClientAddress("192.0.2.1"),
				semconv.ClientPort(1234),
				semconv.URLFull("https://example.com:8080/uri/test"),
				semconv.URLPath("/uri/test"),
				semconv.NetworkProtocolVersion("1.1"),
				semconv.NetworkPeerAddress("192.0.2.1"),
				semconv.NetworkPeerPort(1234),
				semconv.HTTPResponseStatusCode(http.StatusPermanentRedirect),
			}
			assert.Equal(t, wantAttrs, span.Attributes())
			events := span.Events()
			assert.Empty(t, events)

			scope := span.InstrumentationScope()
			assert.Equal(t, otel.OpenTelemetryTracerName, scope.Name)
			assert.Equal(t, otel.Version, scope.Version)

			if assert.Len(t, metrics.ScopeMetrics, 1) {
				sm := metrics.ScopeMetrics[0]
				assert.Equal(t, otel.OpenTelemetryMeterName, sm.Scope.Name)
				assert.Equal(t, otel.Version, sm.Scope.Version)

				if assert.Len(t, sm.Metrics, 3) {
					requestBodySizeMetric := sm.Metrics[0]
					responseBodySizeMetric := sm.Metrics[1]
					requestDurationMetric := sm.Metrics[2]

					assertIntMetric(t, requestBodySizeMetric, httpconv.ServerRequestBodySize{}.Name(), 0, http.StatusPermanentRedirect, nil, true)
					assertIntMetric(t, responseBodySizeMetric, httpconv.ServerResponseBodySize{}.Name(), 62, http.StatusPermanentRedirect, nil, true)
					assertDurationMetric(t, requestDurationMetric, httpconv.ServerRequestDuration{}.Name(), 0, http.StatusPermanentRedirect, nil, true)
				}
			}
		})
	})

	t.Run("filter", func(t *testing.T) {
		filterTrue := func(_ *Request) bool {
			return true
		}
		filterFalse := func(_ *Request) bool {
			return false
		}

		cases := []struct {
			desc     string
			filters  []TraceFilter
			wantSpan bool
		}{
			{
				desc:     "no_filter",
				filters:  nil,
				wantSpan: true,
			},
			{
				desc:     "filter_true",
				filters:  []TraceFilter{filterTrue},
				wantSpan: true,
			},
			{
				desc:     "many_filter_true",
				filters:  []TraceFilter{filterTrue, filterTrue},
				wantSpan: true,
			},
			{
				desc:     "filter_false",
				filters:  []TraceFilter{filterFalse},
				wantSpan: false,
			},
			{
				desc:     "filter_first_false",
				filters:  []TraceFilter{filterFalse, filterTrue},
				wantSpan: false,
			},
			{
				desc:     "filter_second_false",
				filters:  []TraceFilter{filterTrue, filterFalse},
				wantSpan: false,
			},
		}

		for _, c := range cases {
			t.Run(c.desc, func(t *testing.T) {
				spanRecorder, _ := prepareOpenTelemetryTest(t, "/test/url", func(_ *Response, _ *Request) {}, c.filters...)

				spans := spanRecorder.Ended()
				if c.wantSpan {
					require.Len(t, spans, 1)
				} else {
					require.Empty(t, spans)
				}
			})
		}
	})
}

func prepareOpenTelemetryTest(t *testing.T, url string, handler Handler, traceFilters ...TraceFilter) (*tracetest.SpanRecorder, *metricdata.ResourceMetrics) {
	spanRecorder := tracetest.NewSpanRecorder()
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(spanRecorder),
	)
	meterReader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(meterReader))
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	t.Cleanup(func() {
		_ = traceProvider.Shutdown(t.Context())
		_ = meterProvider.Shutdown(t.Context())
		_ = meterReader.Shutdown(t.Context())
	})

	opts := Options{
		OpenTelemetry: OpenTelemetryOptions{
			TracerProvider: traceProvider,
			MeterProvider:  meterProvider,
			Propagators:    propagator,
			TraceFilters:   traceFilters,
		},
		Logger: slog.DiscardLogger(),
	}
	server, err := New(config.LoadDefault(), opts)
	require.NoError(t, err)

	router := server.Router()
	router.Get("/uri/{param}", handler)

	httpRecorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(server.ctx, http.MethodGet, url, bytes.NewReader([]byte{1, 2, 3, 4, 5}))
	request.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	request.Header.Set("tracestate", "congo=t61rcWkgMzE")
	request.Header.Set("baggage", "userId=alice,isProduction=false")
	router.ServeHTTP(httpRecorder, request)

	var metrics metricdata.ResourceMetrics
	err = meterReader.Collect(t.Context(), &metrics)
	require.NoError(t, err)

	return spanRecorder, &metrics
}

func assertIntMetric(t *testing.T, metric metricdata.Metrics, wantKey string, wantValue int64, wantStatus int, err error, skipAttrsCheck bool) {
	assert.Equal(t, string(wantKey), metric.Name)

	requestBodySizeHist := metric.Data.(metricdata.Histogram[int64])
	if !assert.Len(t, requestBodySizeHist.DataPoints, 1) {
		return
	}
	dp := requestBodySizeHist.DataPoints[0]

	if !skipAttrsCheck {
		assertMetricAttrs(t, dp, wantStatus, err)
	}
	assert.Equal(t, uint64(1), dp.Count)

	bounds := []float64{
		0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000,
	}
	assert.Equal(t, bounds, dp.Bounds)

	min, _ := dp.Min.Value()
	max, _ := dp.Max.Value()
	assert.Equal(t, wantValue, min)
	assert.Equal(t, wantValue, max)
	assert.Equal(t, wantValue, dp.Sum)
}

func assertDurationMetric(t *testing.T, metric metricdata.Metrics, wantKey string, wantValue float64, wantStatus int, err error, skipAttrsCheck bool) {
	assert.Equal(t, string(wantKey), metric.Name)

	requestBodySizeHist := metric.Data.(metricdata.Histogram[float64])
	if !assert.Len(t, requestBodySizeHist.DataPoints, 1) {
		return
	}
	dp := requestBodySizeHist.DataPoints[0]

	if !skipAttrsCheck {
		assertMetricAttrs(t, dp, wantStatus, err)
	}
	assert.Equal(t, uint64(1), dp.Count)

	bounds := []float64{
		0.005, 0.01, 0.025, 0.05, 0.075, 0.1,
		0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10,
	}
	assert.Equal(t, bounds, dp.Bounds)

	min, _ := dp.Min.Value()
	max, _ := dp.Max.Value()
	if wantValue == 0 {
		assert.InDelta(t, 0, min, 0)
		assert.InDelta(t, 0, max, 0)
		assert.InDelta(t, 0, dp.Sum, 0)
	} else {
		assert.InDelta(t, wantValue, min, 0.00000001)
		assert.InDelta(t, wantValue, max, 0.00000001)
		assert.InDelta(t, wantValue, dp.Sum, 0.00000001)
	}
}

func assertMetricAttrs[T int64 | float64](t *testing.T, dp metricdata.HistogramDataPoint[T], status int, err error) {
	wantAttrs := []attribute.KeyValue{
		semconv.HTTPRequestMethodGet,
		semconv.HTTPRoute("/uri/{param}"),
		semconv.NetworkProtocolVersion("1.1"),
		semconv.ServerAddress("example.com"),
	}
	if err != nil {
		wantAttrs = append(wantAttrs, semconv.ErrorType(err))
	}
	wantAttrs = append(wantAttrs, semconv.HTTPResponseStatusCode(status))
	set := attribute.NewSet(wantAttrs...)

	assert.Equal(t, set, dp.Attributes)
}
