package ent

import (
	"database/sql"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func (c *Client) DB() *sql.DB {
	var sqlDriver *entsql.Driver
	debugDriver, ok := c.driver.(*dialect.DebugDriver)
	if ok {
		sqlDriver = debugDriver.Driver.(*entsql.Driver)
	} else {
		sqlDriver = c.driver.(*entsql.Driver)
	}
	return sqlDriver.DB()
}
