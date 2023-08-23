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

	xslog "log/slog"

	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"goyave.dev/goyave/v5/slog"
)

var regexGormPath = regexp.MustCompile(`gorm.io/(.*?)@`)

// Logger adapter between `*slog.Logger` and GORM's logger.
type Logger struct {
	slogger       *slog.Logger
	SlowThreshold time.Duration
}

func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	return &newlogger
}

func (l Logger) Info(ctx context.Context, msg string, data ...any) {
	if l.slogger == nil {
		return
	}
	l.slogger.InfoWithSource(ctx, getSourceCaller(), fmt.Sprintf(msg, data...))
}

func (l Logger) Warn(ctx context.Context, msg string, data ...any) {
	if l.slogger == nil {
		return
	}
	l.slogger.WarnWithSource(ctx, getSourceCaller(), fmt.Sprintf(msg, data...))
}

func (l Logger) Error(ctx context.Context, msg string, data ...any) {
	if l.slogger == nil {
		return
	}
	l.slogger.ErrorWithSource(ctx, getSourceCaller(), fmt.Errorf(msg, data...))
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.slogger == nil {
		return
	}

	elapsed := time.Since(begin)

	switch {
	case err != nil && l.slogger.Enabled(ctx, xslog.LevelError) && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		l.slogger.ErrorWithSource(ctx, getSourceCaller(), fmt.Errorf("%s\n"+slog.Reset+slog.Yellow+"[%.3fms] "+slog.Blue+"[rows:%s]"+slog.Reset+" %s", err.Error(), float64(elapsed.Nanoseconds())/1e6, lo.Ternary(rows == -1, "-", strconv.FormatInt(rows, 10)), sql))
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.slogger.Enabled(ctx, xslog.LevelWarn):
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.slogger.WarnWithSource(ctx, getSourceCaller(), fmt.Sprintf("%s\n"+slog.Reset+slog.Red+"[%.3fms] "+slog.Blue+"[rows:%s]"+slog.Reset+" %s", slowLog, float64(elapsed.Nanoseconds())/1e6, lo.Ternary(rows == -1, "-", strconv.FormatInt(rows, 10)), sql))
	case l.slogger.Enabled(ctx, xslog.LevelDebug):
		sql, rows := fc()
		l.slogger.DebugWithSource(ctx, getSourceCaller(), fmt.Sprintf(slog.Yellow+"[%.3fms] "+slog.Blue+"[rows:%s]"+slog.Reset+" %s", float64(elapsed.Nanoseconds())/1e6, lo.Ternary(rows == -1, "-", strconv.FormatInt(rows, 10)), sql))
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
