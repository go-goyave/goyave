package database

import (
	"testing"

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
		Title:   "lorem ipsum",
		Content: "lorem ipsum sit dolor amet",
	}
}

func preparePaginatorTestDB() (*gorm.DB, []*TestArticle) {
	cfg := config.LoadDefault()
	cfg.Set("app.debug", false)
	cfg.Set("database.connection", "sqlite3_paginator_test")
	cfg.Set("database.name", "paginator_test.db")
	cfg.Set("database.options", "mode=memory")
	db, err := New(cfg, nil)
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&TestUser{}, &TestArticle{}); err != nil {
		panic(err)
	}

	author := userGenerator()
	if err := db.Create(author).Error; err != nil {
		panic(err)
	}

	factory := NewFactory(articleGenerator)
	factory.Override(&TestArticle{AuthorID: author.ID})

	articles := factory.Save(db, 11)
	return db, articles
}

func TestPaginator(t *testing.T) {
	RegisterDialect("sqlite3_paginator_test", "file:{name}?{options}", sqlite.Open)
	t.Cleanup(func() {
		mu.Lock()
		delete(dialects, "sqlite3_paginator_test")
		mu.Unlock()
	})

	t.Run("UpdatePageInfo", func(t *testing.T) {
		db, _ := preparePaginatorTestDB()
		articles := []*TestArticle{}
		p := NewPaginator(db, 2, 5, &articles)

		assert.Equal(t, db, p.DB)
		assert.Equal(t, 2, p.CurrentPage)
		assert.Equal(t, 5, p.PageSize)
		assert.Equal(t, &articles, p.Records)

		err := p.UpdatePageInfo()
		require.NoError(t, err)

		assert.Equal(t, int64(11), p.Total)
		assert.Equal(t, int64(3), p.MaxPage)
		assert.True(t, p.loadedPageInfo)
	})

	t.Run("Find", func(t *testing.T) {
		db, srcArticles := preparePaginatorTestDB()
		articles := []*TestArticle{}
		p := NewPaginator(db, 2, 5, &articles)

		assert.Equal(t, db, p.DB)
		assert.Equal(t, 2, p.CurrentPage)
		assert.Equal(t, 5, p.PageSize)
		assert.Equal(t, &articles, p.Records)

		err := p.Find()
		require.NoError(t, err)

		assert.Equal(t, int64(11), p.Total)
		assert.Equal(t, int64(3), p.MaxPage)
		assert.True(t, p.loadedPageInfo)
		assert.Equal(t, srcArticles[5:10], *p.Records)
	})

	t.Run("Find_no_record", func(t *testing.T) {
		cfg := config.LoadDefault()
		cfg.Set("app.debug", false)
		cfg.Set("database.connection", "sqlite3_paginator_test")
		cfg.Set("database.name", "paginator_test.db")
		cfg.Set("database.options", "mode=memory")
		db, err := New(cfg, nil)
		if err != nil {
			panic(err)
		}

		if err := db.AutoMigrate(&TestUser{}, &TestArticle{}); err != nil {
			panic(err)
		}

		articles := []*TestArticle{}
		p := NewPaginator(db, 2, 5, &articles)

		err = p.Find()
		require.NoError(t, err)

		assert.Equal(t, int64(0), p.Total)
		assert.Equal(t, int64(1), p.MaxPage)
		assert.True(t, p.loadedPageInfo)
		assert.Empty(t, *p.Records)
	})

	t.Run("UpdatePageInfo_error", func(t *testing.T) {
		db, _ := preparePaginatorTestDB()
		articles := []*TestArticle{}

		db = db.Where("not_a_column", 1)
		p := NewPaginator(db, 2, 5, &articles)

		err := p.Find() // updatePageInfo is called because the page info is not called yet
		require.Error(t, err)
		assert.False(t, p.loadedPageInfo)
	})

	t.Run("Find_error", func(t *testing.T) {
		db, _ := preparePaginatorTestDB()
		articles := []*TestArticle{}

		db = db.Where("not_a_column", 1)
		p := NewPaginator(db, 2, 5, &articles)
		p.loadedPageInfo = true // Let's assume the page info has already been loaded

		err := p.Find()
		require.Error(t, err)
		assert.False(t, p.loadedPageInfo) // Page info invalidated
	})

	t.Run("select_where_preload", func(t *testing.T) {
		db, _ := preparePaginatorTestDB()
		articles := []*TestArticle{}

		db = db.Select("id", "title", "author_id").Where("id > ?", 9).Preload("Author")
		p := NewPaginator(db, 1, 5, &articles)

		err := p.Find()
		require.NoError(t, err)

		assert.Equal(t, int64(2), p.Total)
		assert.Equal(t, int64(1), p.MaxPage)
		assert.True(t, p.loadedPageInfo)

		author := userGenerator()
		author.ID = 1
		expected := []*TestArticle{
			{ID: 10, Title: "lorem ipsum", AuthorID: 1, Author: author},
			{ID: 11, Title: "lorem ipsum", AuthorID: 1, Author: author},
		}
		assert.Equal(t, expected, *p.Records)
	})

	t.Run("Raw", func(t *testing.T) {
		db, _ := preparePaginatorTestDB()
		articles := []*TestArticle{}
		p := NewPaginator(db, 1, 5, &articles)

		query := `SELECT id, title FROM test_articles WHERE id > ?`
		queryVars := []any{9}
		countQuery := `SELECT COUNT(*) FROM test_articles WHERE id > ?`
		assert.Equal(t, p, p.Raw(query, queryVars, countQuery, queryVars))

		err := p.Find()
		require.NoError(t, err)

		assert.Equal(t, int64(2), p.Total)
		assert.Equal(t, int64(1), p.MaxPage)
		assert.True(t, p.loadedPageInfo)

		expected := []*TestArticle{
			{ID: 10, Title: "lorem ipsum"},
			{ID: 11, Title: "lorem ipsum"},
		}
		assert.Equal(t, expected, *p.Records)

		// Get page 2 (no results expected)
		articles = []*TestArticle{}
		p = NewPaginator(db, 2, 5, &articles)
		p.Raw(query, queryVars, countQuery, queryVars)
		err = p.Find()
		require.NoError(t, err)
		assert.Equal(t, int64(2), p.Total)
		assert.Equal(t, int64(1), p.MaxPage)
		assert.True(t, p.loadedPageInfo)
		assert.Empty(t, *p.Records)
	})
}
