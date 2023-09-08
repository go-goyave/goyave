package database

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	stdslog "log/slog"

	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"goyave.dev/goyave/v5/slog"
)

var regexGormPath = regexp.MustCompile(`gorm.io/(.*?)@`)

// Logger adapter between `*slog.Logger` and GORM's logger.
type Logger struct {
	slogger func() *slog.Logger

	// SlowThreshold defines the minimum query execution time to be considered "slow".
	// If a query takes more time than `SlowThreshold`, the query will be logged at the WARN level.
	// If 0, disables query execution time checking.
	SlowThreshold time.Duration
}

// NewLogger create a new `Logger` adapter between GORM and `*slog.Logger`.
// Use a `SlowThreshold` of 200ms.
func NewLogger(slogger func() *slog.Logger) *Logger {
	return &Logger{
		slogger:       slogger,
		SlowThreshold: 200 * time.Millisecond,
	}
}

// LogMode returns a copy of this logger. The level argument actually has
// no effect as it is handled by the underlying `*slog.Logger`.
func (l *Logger) LogMode(_ logger.LogLevel) logger.Interface {
	newlogger := *l
	return &newlogger
}

// Info logs at `LevelInfo`.
func (l Logger) Info(ctx context.Context, msg string, data ...any) {
	if l.slogger == nil {
		return
	}
	l.slogger().InfoWithSource(ctx, getSourceCaller(), fmt.Sprintf(msg, data...))
}

// Warn logs at `LevelWarn`.
func (l Logger) Warn(ctx context.Context, msg string, data ...any) {
	if l.slogger == nil {
		return
	}
	l.slogger().WarnWithSource(ctx, getSourceCaller(), fmt.Sprintf(msg, data...))
}

// Error logs at `LevelError`.
func (l Logger) Error(ctx context.Context, msg string, data ...any) {
	if l.slogger == nil {
		return
	}
	l.slogger().ErrorWithSource(ctx, getSourceCaller(), fmt.Errorf(msg, data...))
}

// Trace SQL logs at
//   - `LevelDebug`
//   - `LevelWarn` if the query is slow
//   - `LevelError` if the given error is not nil
func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.slogger == nil {
		return
	}

	elapsed := time.Since(begin)

	switch {
	case err != nil && l.slogger().Enabled(ctx, stdslog.LevelError) && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		l.slogger().ErrorWithSource(ctx, getSourceCaller(), fmt.Errorf("%s\n"+slog.Reset+slog.Yellow+"[%.3fms] "+slog.Blue+"[rows:%s]"+slog.Reset+" %s", err.Error(), float64(elapsed.Nanoseconds())/1e6, lo.Ternary(rows == -1, "-", strconv.FormatInt(rows, 10)), sql))
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.slogger().Enabled(ctx, stdslog.LevelWarn):
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.slogger().WarnWithSource(ctx, getSourceCaller(), fmt.Sprintf("%s\n"+slog.Reset+slog.Red+"[%.3fms] "+slog.Blue+"[rows:%s]"+slog.Reset+" %s", slowLog, float64(elapsed.Nanoseconds())/1e6, lo.Ternary(rows == -1, "-", strconv.FormatInt(rows, 10)), sql))
	case l.slogger().Enabled(ctx, stdslog.LevelDebug):
		sql, rows := fc()
		l.slogger().DebugWithSource(ctx, getSourceCaller(), fmt.Sprintf(slog.Yellow+"[%.3fms] "+slog.Blue+"[rows:%s]"+slog.Reset+" %s", float64(elapsed.Nanoseconds())/1e6, lo.Ternary(rows == -1, "-", strconv.FormatInt(rows, 10)), sql))
	}
}

func getSourceCaller() uintptr {
	// function copied from gorm/utils/utils.go
	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		pc, file, _, ok := runtime.Caller(i)
		if ok && (!regexGormPath.MatchString(file) || strings.HasSuffix(file, "_test.go")) {
			return pc
		}
	}

	return 0
}
