package database

import (
	"database/sql/driver"
	"fmt"
	"math/rand/v2"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
)

type TestArticle struct {
	Author *TestUser `gorm:"foreignKey:AuthorID"`

	Title    string `gorm:"type:varchar(255)"`
	Content  string `gorm:"type:text"`
	ID       uint   `gorm:"primaryKey"`
	AuthorID uint
}

func articleGenerator() *TestArticle {
	return &TestArticle{
		ID:       rand.UintN(9999),
		Title:    "lorem ipsum",
		Content:  "lorem ipsum sit dolor amet",
		AuthorID: rand.UintN(100),
	}
}

func generateRows(count int) (*sqlmock.Rows, []*TestArticle) {
	articles := make([]*TestArticle, 0, count)
	rows := sqlmock.NewRows([]string{"id", "title", "content", "author_id"})
	for range count {
		article := articleGenerator()
		articles = append(articles, article)
		rows = rows.AddRow(article.ID, article.Title, article.Content, article.AuthorID)
	}
	return rows, articles
}

func generateRowsWithAuthor(count int) (*sqlmock.Rows, *sqlmock.Rows, []*TestArticle) {
	articles := make([]*TestArticle, 0, count)
	articleRows := sqlmock.NewRows([]string{"id", "title", "content", "author_id"})
	authorRows := sqlmock.NewRows([]string{"id", "name", "email"})
	for i := range count {
		author := userGenerator()
		author.ID = uint(i + 1)
		article := articleGenerator()
		article.Author = author
		article.AuthorID = author.ID
		articles = append(articles, article)
		articleRows = articleRows.AddRow(article.ID, article.Title, article.Content, article.AuthorID)
		authorRows = authorRows.AddRow(author.ID, author.Name, author.Email)
	}
	return articleRows, authorRows, articles
}

func preparePaginatorTest(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	cfg := &config.DatabaseConnection{
		Dialect:            "sqlmock",
		DatabaseName:       "paginator_test.db",
		MaxIdleConnections: 1,
		GORM:               config.GORM{}, // Disabling PrepareStmt is important to avoid errors caused by mock
	}

	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	dialector := &sqlite.Dialector{
		DriverName: "sqlite3_timeout_test",
		DSN:        fmt.Sprintf("file:%s?%s", cfg.DatabaseName, cfg.Options),
		Conn:       mockDB,
	}

	t.Cleanup(func() {
		mock.ExpectClose()
		assert.NoError(t, mockDB.Close())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// The SQLite dialector selects the sqlite version first to know which callback clauses it can use.
	mock.ExpectQuery(regexp.QuoteMeta(`select sqlite_version()`)).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("3.53.4"))

	db, err := NewFromDialector(cfg, nil, dialector)
	if err != nil {
		require.NoError(t, err)
	}
	return db, mock
}

func TestPaginator(t *testing.T) {
	t.Run("UpdatePageInfo", func(t *testing.T) {
		db, mock := preparePaginatorTest(t)
		articles := []*TestArticle{}
		p := NewPaginator(db, 2, 5, &articles)

		assert.Equal(t, db, p.DB)
		assert.Equal(t, 2, p.CurrentPage)
		assert.Equal(t, 5, p.PageSize)
		assert.Equal(t, &articles, p.Records)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `test_articles`")).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(11))

		err := p.UpdatePageInfo()
		require.NoError(t, err)

		assert.Equal(t, int64(11), p.Total)
		assert.Equal(t, int64(3), p.MaxPage)
		assert.True(t, p.loadedPageInfo)
	})

	t.Run("Find", func(t *testing.T) {
		db, mock := preparePaginatorTest(t)
		articles := []*TestArticle{}
		p := NewPaginator(db, 3, 5, &articles)

		assert.Equal(t, db, p.DB)
		assert.Equal(t, 3, p.CurrentPage)
		assert.Equal(t, 5, p.PageSize)
		assert.Equal(t, &articles, p.Records)

		rows, want := generateRows(2)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `test_articles`")).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(12))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `test_articles` LIMIT 5 OFFSET 10")).
			WillReturnRows(rows)
		mock.ExpectCommit()

		err := p.Find()
		require.NoError(t, err)

		assert.Equal(t, int64(12), p.Total)
		assert.Equal(t, int64(3), p.MaxPage)
		assert.True(t, p.loadedPageInfo)
		assert.Equal(t, want, *p.Records)
		assert.Equal(t, want, articles)
	})

	t.Run("Find_no_record", func(t *testing.T) {
		db, mock := preparePaginatorTest(t)
		articles := []*TestArticle{}
		p := NewPaginator(db, 3, 5, &articles)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `test_articles`")).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(0))
		mock.ExpectCommit()

		err := p.Find()
		require.NoError(t, err)

		assert.Equal(t, int64(0), p.Total)
		assert.Equal(t, int64(1), p.MaxPage)
		assert.True(t, p.loadedPageInfo)
		assert.Empty(t, *p.Records)
		assert.Empty(t, articles)
	})

	t.Run("UpdatePageInfo_error", func(t *testing.T) {
		db, mock := preparePaginatorTest(t)
		articles := []*TestArticle{}

		db = db.Where("not_a_column", 1)
		p := NewPaginator(db, 2, 5, &articles)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `test_articles` WHERE `not_a_column` = ?")).
			WithArgs(1).
			WillReturnError(fmt.Errorf("not_a_column does not exist on table test_articles"))
		mock.ExpectRollback()

		err := p.Find() // updatePageInfo is called because the page info is not called yet
		require.Error(t, err)
		assert.False(t, p.loadedPageInfo)
	})

	t.Run("Find_error", func(t *testing.T) {
		db, mock := preparePaginatorTest(t)
		articles := []*TestArticle{}

		db = db.Where("not_a_column", 1)
		p := NewPaginator(db, 2, 5, &articles)
		p.loadedPageInfo = true // Let's assume the page info has already been loaded
		p.Total = 11

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `test_articles`  WHERE `not_a_column` = ? LIMIT 5 OFFSET 5")).
			WithArgs(1).
			WillReturnError(fmt.Errorf("not_a_column does not exist on table test_articles"))
		mock.ExpectRollback()

		err := p.Find()
		require.Error(t, err)
		assert.False(t, p.loadedPageInfo) // Page info invalidated
	})

	t.Run("select_where_preload", func(t *testing.T) {
		db, mock := preparePaginatorTest(t)
		articles := []*TestArticle{}

		db = db.Select("id", "title", "author_id").Where("id > ?", 9).Preload("Author")
		p := NewPaginator(db, 1, 5, &articles)

		articleRows, authorRows, want := generateRowsWithAuthor(3)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `test_articles` WHERE id > ?")).
			WithArgs(9).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(3))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT `id`,`title`,`author_id` FROM `test_articles` WHERE id > ? LIMIT 5")).
			WithArgs(9).
			WillReturnRows(articleRows)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `test_users` WHERE `test_users`.`id` IN (?,?,?)")).
			WithArgs(lo.Map(want, func(a *TestArticle, _ int) driver.Value { return a.AuthorID })...).
			WillReturnRows(authorRows)
		mock.ExpectCommit()

		err := p.Find()
		require.NoError(t, err)

		assert.Equal(t, int64(3), p.Total)
		assert.Equal(t, int64(1), p.MaxPage)
		assert.True(t, p.loadedPageInfo)

		assert.Equal(t, want, *p.Records)
		assert.Equal(t, want, articles)
	})

	t.Run("Raw", func(t *testing.T) {
		db, mock := preparePaginatorTest(t)
		articles := []*TestArticle{}
		p := NewPaginator(db, 1, 5, &articles)

		query := `SELECT id, title FROM test_articles WHERE id > ?`
		queryVars := []any{9}
		countQuery := `SELECT count(*) FROM test_articles WHERE id > ?`
		assert.Equal(t, p, p.Raw(query, queryVars, countQuery, queryVars))

		rows, want := generateRows(2)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(countQuery)).
			WithArgs(9).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(2))
		mock.ExpectQuery(regexp.QuoteMeta(query+" LIMIT ?")).
			WithArgs(9, 5).
			WillReturnRows(rows)
		mock.ExpectCommit()

		err := p.Find()
		require.NoError(t, err)

		assert.Equal(t, int64(2), p.Total)
		assert.Equal(t, int64(1), p.MaxPage)
		assert.True(t, p.loadedPageInfo)

		assert.Equal(t, want, *p.Records)
		assert.Equal(t, want, articles)

		// Get page 2 (no results expected)
		articles = []*TestArticle{}
		p = NewPaginator(db, 2, 5, &articles)
		p.Raw(query, queryVars, countQuery, queryVars)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(countQuery)).
			WithArgs(9).
			WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(2))
		mock.ExpectQuery(regexp.QuoteMeta(query+" LIMIT ? OFFSET ?")).
			WithArgs(9, 5, 5).
			WillReturnRows(sqlmock.NewRows([]string{}))
		mock.ExpectCommit()

		err = p.Find()
		require.NoError(t, err)
		assert.Equal(t, int64(2), p.Total)
		assert.Equal(t, int64(1), p.MaxPage)
		assert.True(t, p.loadedPageInfo)
		assert.Empty(t, *p.Records)
		assert.Empty(t, articles)
	})
}
