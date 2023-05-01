package main

import (
	"encoding/json"
	"fmt"

	"github.com/itchyny/gojq"
	"go.riyazali.net/sqlite"
)

//const (
//	JEACH_KEY = iota
//	JEACH_VALUE
//	JEACH_TYPE
//	JEACH_ATOM
//	JEACH_ID
//	JEACH_PARENT
//	JEACH_FULLKEY
//	JEACH_PATH
//)

type JqEachModule struct{}

func (m *JqEachModule) Connect(_ *sqlite.Conn, _ []string, declare func(string) error) (sqlite.VirtualTable, error) {
	return &JqEachTable{}, declare("CREATE TABLE jq_each(value, json hidden, query hidden)")
}

type JqEachTable struct{}

func (s *JqEachTable) Open() (sqlite.VirtualCursor, error) {
	return &JqEachCursor{}, nil
}

func (s *JqEachTable) Disconnect() error { return s.Destroy() }
func (s *JqEachTable) Destroy() error    { return nil }

type JqEachCursor struct {
	rowid int64
	query *gojq.Query
	iter  gojq.Iter
	value *interface{}
}

func (s *JqEachTable) BestIndex(input *sqlite.IndexInfoInput) (*sqlite.IndexInfoOutput, error) {
	var args = 0

	var output = &sqlite.IndexInfoOutput{
		ConstraintUsage: make([]*sqlite.ConstraintUsage, len(input.Constraints)),
		EstimatedCost:   1000,
	}

	for j, con := range input.Constraints {
		if con.ColumnIndex < 1 {
			continue
		}

		/* We need all of the constraints to be usable */
		if !con.Usable {
			return nil, sqlite.SQLITE_CONSTRAINT
		}

		args += 1
		output.ConstraintUsage[j] = &sqlite.ConstraintUsage{ArgvIndex: args, Omit: true}
	}

	return output, nil
}

func (c *JqEachCursor) Rowid() (int64, error) {
	return c.rowid, nil
}

func (c *JqEachCursor) Filter(idxNum int, _ string, values ...sqlite.Value) error {
	if len(values) == 0 {
		c.rowid = -1
		return sqlite.SQLITE_OK
	}

	for _, v := range values { // if any of the constraints have a NULL value, then return no rows.
		if v.Type() == sqlite.SQLITE_NULL {
			c.rowid = -1
			return sqlite.SQLITE_OK
		}
	}

	var val interface{}
	err := json.Unmarshal(values[0].Blob(), &val)

	if err != nil {
		err = fmt.Errorf("error parsing JSON data: %v", err)
		return err
	}

	query, err := gojq.Parse(values[1].Text())

	if err != nil {
		err = fmt.Errorf("error parsing JQ query: %v", err)
		return err
	}

	c.rowid = 0
	c.query = query
	c.iter = query.Run(val)

	/* The iterator won't be at the first row, so manually move it */
	return c.Next()
}

func (c *JqEachCursor) Next() error {
	v, ok := c.iter.Next()

	if !ok {
		c.rowid = -1
		return nil
	}

	if err, ok := v.(error); ok {
		err = fmt.Errorf("error creating result: %v", err)
		return err
	}

	c.rowid += 1
	c.value = &v
	return nil
}

func (c *JqEachCursor) Column(ctx *sqlite.VirtualTableContext, i int) error {
	if c.value == nil {
		ctx.ResultNull()
		return sqlite.SQLITE_OK
	}

	v := *c.value

	switch v := v.(type) {
	case bool:
		if v {
			ctx.ResultInt(1)
		} else {
			ctx.ResultInt(0)
		}
	case int:
		ctx.ResultInt(v)
	case int64:
		ctx.ResultInt64(v)
	case float64:
		ctx.ResultFloat(v)
	case string:
		ctx.ResultText(v)
	default:
		tmp, err := json.Marshal(v)

		if err != nil {
			err = fmt.Errorf("error marshalling result data: %v", err)
			ctx.ResultError(err)
			return nil
		}

		ctx.ResultBlob(tmp)
	}

	return nil
}

func (c *JqEachCursor) Eof() bool {
	return c.rowid == -1
}

func (c *JqEachCursor) Close() error { return nil }
