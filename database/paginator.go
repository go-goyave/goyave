package database

import (
	"math"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Paginator structure containing pagination information and result records.
// Can be sent to the client directly.
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
//	result := paginator.Find()
//	if response.WriteDBError(result.Error) {
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

// UpdatePageInfo executes count request to calculate the `Total` and `MaxPage`.
// Returns the executed statement, which may contain an error.
func (p *Paginator[T]) UpdatePageInfo() *gorm.DB {
	count := int64(0)
	db := p.DB.Session(&gorm.Session{})
	prevPreloads := db.Statement.Preloads
	if len(prevPreloads) > 0 {
		db.Statement.Preloads = map[string][]any{}
		defer func() {
			db.Statement.Preloads = prevPreloads
		}()
	}
	prevSelects := db.Statement.Selects
	if len(prevSelects) > 0 {
		db.Statement.Selects = []string{}
		defer func() {
			db.Statement.Selects = prevSelects
		}()
	}

	var res *gorm.DB
	if p.rawCountQuery != "" {
		res = db.Raw(p.rawCountQuery, p.rawCountQueryVars...).Scan(&count)
	} else {
		res = db.Model(p.Records).Count(&count)
	}
	if res.Error != nil {
		return res
	}
	p.Total = count
	p.MaxPage = int64(math.Ceil(float64(count) / float64(p.PageSize)))
	if p.MaxPage == 0 {
		p.MaxPage = 1
	}
	p.loadedPageInfo = true
	return res
}

// Find requests page information (total records and max page) and
// executes the transaction. The Paginate struct is updated automatically, as
// well as the destination slice given in `NewPaginator()`.
//
// Returns the executed statement, which may contain an error.
func (p *Paginator[T]) Find() *gorm.DB {
	if !p.loadedPageInfo {
		res := p.UpdatePageInfo()
		if res.Error != nil {
			return res
		}
	}
	if p.rawQuery != "" {
		return p.rawStatement().Scan(p.Records)
	}
	return p.DB.Scopes(paginateScope(p.CurrentPage, p.PageSize)).Find(p.Records)
}

func (p *Paginator[T]) rawStatement() *gorm.DB {
	offset := (p.CurrentPage - 1) * p.PageSize
	db := p.DB.Raw(p.rawQuery, p.rawQueryVars...)
	pageSize := p.PageSize
	db.Statement.SQL.WriteString(" ")
	clause.Limit{Limit: &pageSize, Offset: offset}.Build(db.Statement)
	return db
}
