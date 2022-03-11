package database

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"goyave.dev/goyave/v4/config"
)

type PaginatorTestSuite struct {
	suite.Suite
	previousEnv string
}

func (suite *PaginatorTestSuite) SetupSuite() {
	if _, ok := dialects["mysql"]; !ok {
		RegisterDialect("mysql", "{username}:{password}@({host}:{port})/{name}?{options}", mysql.Open)
	}
	suite.previousEnv = os.Getenv("GOYAVE_ENV")
	os.Setenv("GOYAVE_ENV", "test")
	if err := config.Load(); err != nil {
		suite.FailNow(err.Error())
	}

	db := GetConnection()
	if err := db.AutoMigrate(&User{}); err != nil {
		panic(err)
	}
}

func (suite *PaginatorTestSuite) TestPaginator() {
	// Generate records
	const userCount = 21
	users := make([]User, 0, userCount)
	for i := 0; i < userCount; i++ {
		users = append(users, User{"John Doe", "johndoe@example.org", 0})
	}

	db := GetConnection()
	if err := db.Create(users).Error; err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{}).Error; err != nil {
			panic(err)
		}
	}()

	results := []User{}
	paginator := NewPaginator(db, 1, 10, &results)
	suite.Equal(1, paginator.CurrentPage)
	suite.Equal(10, paginator.PageSize)
	suite.Equal(db, paginator.DB)
	res := paginator.Find()
	suite.Nil(res.Error)
	if res.Error == nil {
		suite.Len(results, 10)
		suite.Equal(int64(userCount), paginator.Total)
		suite.Equal(int64(3), paginator.MaxPage)
	}

	results = []User{}
	paginator = NewPaginator(db, 3, 10, &results)
	res = paginator.Find()
	suite.Nil(res.Error)
	if res.Error == nil {
		suite.Len(results, 1)
	}

	results = []User{}
	paginator = NewPaginator(db, 4, 10, &results)
	res = paginator.Find()
	suite.Nil(res.Error)
	if res.Error == nil {
		suite.Empty(results)
	}

	// MaxPage = 1 is there is no record
	if err := db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{}).Error; err != nil {
		panic(err)
	}
	results = []User{}
	paginator = NewPaginator(db, 1, 10, &results)
	res = paginator.Find()
	suite.Nil(res.Error)
	if res.Error == nil {
		suite.Empty(results)
		suite.Equal(int64(1), paginator.MaxPage)
	}
}

func (suite *PaginatorTestSuite) TestPaginatorNoRecord() {
	results := []User{}
	db := GetConnection()
	paginator := NewPaginator(db, 1, 10, &results)
	res := paginator.Find()
	suite.Nil(res.Error)
	if res.Error == nil {
		suite.Empty(results)
		suite.Equal(int64(1), paginator.MaxPage)
	}
}

func (suite *PaginatorTestSuite) TestPaginatorWithWhereClause() {
	// Generate records
	const userCount = 10
	users := make([]User, 0, userCount)
	for i := 0; i < userCount; i++ {
		users = append(users, User{strconv.Itoa(i), "johndoe@example.org", 0})
	}

	db := GetConnection()
	if err := db.Create(users).Error; err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{}).Error; err != nil {
			panic(err)
		}
	}()

	results := []User{}
	db = db.Where("name = ?", "1")
	paginator := NewPaginator(db, 1, 10, &results)
	res := paginator.Find()
	if suite.Nil(res.Error) {
		suite.Len(results, 1)
		suite.Equal(int64(1), paginator.Total)
		suite.Equal(int64(1), paginator.MaxPage)
	}
}

func (suite *PaginatorTestSuite) TestPaginatorRawQuery() {
	// Generate records
	const userCount = 10
	users := make([]User, 0, userCount)
	for i := 0; i < userCount; i++ {
		users = append(users, User{strconv.Itoa(i), "johndoe@example.org", 0})
	}

	db := GetConnection()
	if err := db.Create(users).Error; err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{}).Error; err != nil {
			panic(err)
		}
	}()

	results := []User{}
	vars := []interface{}{0}
	paginator := NewPaginator(db, 1, 5, &results).Raw("SELECT * FROM users WHERE id > ?", vars, "SELECT COUNT(*) FROM users WHERE id > ?", vars)
	res := paginator.Find()
	if suite.Nil(res.Error) {
		suite.Len(results, 5)
		suite.Equal(int64(10), paginator.Total)
		suite.Equal(int64(2), paginator.MaxPage)
	}
}

func (suite *PaginatorTestSuite) TestPaginatorRemovePreloads() {
	// Preloads should be removed for the count query.
	// Generate records
	const userCount = 10
	users := make([]User, 0, userCount)
	for i := 0; i < userCount; i++ {
		users = append(users, User{strconv.Itoa(i), "johndoe@example.org", 0})
	}

	db := GetConnection()
	if err := db.Create(users).Error; err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&User{}).Error; err != nil {
			panic(err)
		}
	}()

	results := []User{}
	db = db.Preload("relation")
	paginator := NewPaginator(db, 1, 5, &results)
	suite.NotPanics(func() {
		paginator.UpdatePageInfo()
		suite.Equal(int64(10), paginator.Total)
		suite.Equal(int64(2), paginator.MaxPage)
	})
}

func (suite *PaginatorTestSuite) TestCountError() {
	db := GetConnection().Table("not a table")
	paginator := NewPaginator(db, 1, 10, []interface{}{})
	suite.Panics(func() {
		paginator.UpdatePageInfo()
	})
}

func (suite *PaginatorTestSuite) TearDownAllSuite() {
	defer os.Setenv("GOYAVE_ENV", suite.previousEnv)
	db := GetConnection()
	if err := db.Migrator().DropTable(&User{}).Error; err != nil {
		panic(err)
	}
}

func TestPaginatorTestSuite(t *testing.T) {
	suite.Run(t, new(PaginatorTestSuite))
}
