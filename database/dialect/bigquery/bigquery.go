package bigquery

import (
	"gorm.io/driver/bigquery"
	"goyave.dev/goyave/v5/database"
)

func init() {
	// location is optional for BigQuery
	// possible name values = ["{projectID}/{location}/{dataSet}", "{projectID}/{dataSet}"]
	database.RegisterDialect("bigquery", "bigquery://{name}?{options}", bigquery.Open)
}
