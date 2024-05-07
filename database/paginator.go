package database

import (
	"math"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"goyave.dev/goyave/v5/util/errors"
)

// Paginator structure containing pagination information and result records.
type Paginator[T any] struct {
	DB *gorm.DB `json:"-"`

	Records *[]T `json:"records"`

	rawQuery          string
	rawQueryVars      []any
	rawCountQuery     string
	rawCountQueryVars []any

	MaxPage     int64 `json:"maxPage"`
	Total       int64 `json:"total"`
	PageSize    int   `json:"pageSize"`
	CurrentPage int   `json:"currentPage"`

	loadedPageInfo bool
}

// PaginatorDTO structure sent to clients as a response.
type PaginatorDTO[T any] struct {
	Records     []T   `json:"records"`
	MaxPage     int64 `json:"maxPage"`
	Total       int64 `json:"total"`
	PageSize    int   `json:"pageSize"`
	CurrentPage int   `json:"currentPage"`
}

func paginateScope(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// NewPaginator create a new Paginator.
//
// Given DB transaction can contain clauses already, such as WHERE, if you want to
// filter results.
//
//	articles := []model.Article{}
//	tx := db.Where("title LIKE ?", "%"+sqlutil.EscapeLike(search)+"%")
//	paginator := database.NewPaginator(tx, page, pageSize, &articles)
//	err := paginator.Find()
//	if response.WriteDBError(err) {
//		return
//	}
//	response.JSON(http.StatusOK, paginator)
func NewPaginator[T any](db *gorm.DB, page, pageSize int, dest *[]T) *Paginator[T] {
	return &Paginator[T]{
		DB:          db,
		CurrentPage: page,
		PageSize:    pageSize,
		Records:     dest,
	}
}

// Raw set a raw SQL query and count query.
// The Paginator will execute the raw queries instead of automatically creating them.
// The raw query should not contain the "LIMIT" and "OFFSET" clauses, they will be added automatically.
// The count query should return a single number (`COUNT(*)` for example).
func (p *Paginator[T]) Raw(query string, vars []any, countQuery string, countVars []any) *Paginator[T] {
	p.rawQuery = query
	p.rawQueryVars = vars
	p.rawCountQuery = countQuery
	p.rawCountQueryVars = countVars
	return p
}

func (p *Paginator[T]) updatePageInfo(db *gorm.DB) error {
	count := int64(0)
	db = db.Session(&gorm.Session{Initialized: true})
	if len(db.Statement.Preloads) > 0 {
		db.Statement.Preloads = map[string][]any{}
	}
	if len(db.Statement.Selects) > 0 {
		db.Statement.Selects = []string{}
	}

	var res *gorm.DB
	if p.rawCountQuery != "" {
		res = db.Raw(p.rawCountQuery, p.rawCountQueryVars...).Scan(&count)
	} else {
		res = db.Model(p.Records).Count(&count)
	}
	if res.Error != nil {
		return errors.New(res.Error)
	}
	p.Total = count
	p.MaxPage = int64(math.Ceil(float64(count) / float64(p.PageSize)))
	if p.MaxPage == 0 {
		p.MaxPage = 1
	}
	p.loadedPageInfo = true
	return nil
}

// UpdatePageInfo executes count request to calculate the `Total` and `MaxPage`.
// When calling this function manually, it is advised to use a transaction that is calling
// `Find()` too, to avoid inconsistencies.
func (p *Paginator[T]) UpdatePageInfo() error {
	return p.updatePageInfo(p.DB)
}

// Find requests page information (total records and max page) if not already fetched using
// `UpdatePageInfo()` and executes the query. The `Paginator` struct is updated automatically,
// as well as the destination slice given in `NewPaginator()`.
//
// The two queries are executed inside a transaction.
func (p *Paginator[T]) Find() error {
	return p.DB.Session(&gorm.Session{}).Transaction(func(tx *gorm.DB) error {
		if !p.loadedPageInfo {
			err := p.updatePageInfo(tx)
			if err != nil {
				return errors.New(err)
			}
		}

		if p.rawQuery != "" {
			p.DB = p.rawStatement(tx).Scan(p.Records)
		} else {
			p.DB = tx.Scopes(paginateScope(p.CurrentPage, p.PageSize)).Find(p.Records)
		}
		if p.DB.Error != nil {
			p.loadedPageInfo = false // Invalidate previous page info.
			return errors.New(p.DB.Error)
		}
		return nil
	})
}

func (p *Paginator[T]) rawStatement(tx *gorm.DB) *gorm.DB {
	offset := (p.CurrentPage - 1) * p.PageSize
	rawStatement := tx.Raw(p.rawQuery, p.rawQueryVars...)
	pageSize := p.PageSize
	rawStatement.Statement.SQL.WriteString(" ")
	clause.Limit{Limit: &pageSize, Offset: offset}.Build(rawStatement.Statement)
	return rawStatement
}
