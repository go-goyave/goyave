package database

import (
	"math"

	"gorm.io/gorm"
)

// Paginator structure containing pagination information and result records.
// Can be sent to the client directly.
type Paginator struct {
	DB *gorm.DB `json:"-"`

	Records     interface{} `json:"records"`
	MaxPage     int64       `json:"maxPage"`
	Total       int64       `json:"total"`
	PageSize    int         `json:"pageSize"`
	CurrentPage int         `json:"currentPage"`

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
//  articles := []model.Article{}
//  tx := database.Conn().Where("title LIKE ?", "%"+sqlutil.EscapeLike(search)+"%")
//  paginator := database.NewPaginator(tx, page, pageSize, &articles)
//  result := paginator.Find()
//  if response.HandleDatabaseError(result) {
//      response.JSON(http.StatusOK, paginator)
//  }
//
func NewPaginator(db *gorm.DB, page, pageSize int, dest interface{}) *Paginator {
	return &Paginator{
		DB:          db,
		CurrentPage: page,
		PageSize:    pageSize,
		Records:     dest,
	}
}

// UpdatePageInfo executes count request to calculate the `Total` and `MaxPage`.
func (p *Paginator) UpdatePageInfo() {
	count := int64(0)
	if err := p.DB.Session(&gorm.Session{}).Model(p.Records).Count(&count).Error; err != nil {
		panic(err)
	}
	p.Total = count
	p.MaxPage = int64(math.Ceil(float64(count) / float64(p.PageSize)))
	if p.MaxPage == 0 {
		p.MaxPage = 1
	}
	p.loadedPageInfo = true
}

// Find requests page information (total records and max page) and
// executes the transaction. The Paginate struct is updated automatically, as
// well as the destination slice given in NewPaginator().
func (p *Paginator) Find() *gorm.DB {
	if !p.loadedPageInfo {
		p.UpdatePageInfo()
	}
	return p.DB.Scopes(paginateScope(p.CurrentPage, p.PageSize)).Find(p.Records)
}
