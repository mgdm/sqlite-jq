package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/itchyny/gojq"
	"go.riyazali.net/sqlite"
)

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

	for _, v := range values {
		if v.Type() == sqlite.SQLITE_NULL {
			c.rowid = -1
			return sqlite.SQLITE_OK
		}
	}

	var val interface{}
	dec := json.NewDecoder(bytes.NewReader(values[0].Blob()))
	dec.UseNumber()
	if err := dec.Decode(&val); err != nil {
		return fmt.Errorf("error parsing JSON data: %w", err)
	}

	code, err := compileQuery(values[1].Text())
	if err != nil {
		return fmt.Errorf("error parsing JQ query: %w", err)
	}

	c.rowid = 0
	c.iter = code.Run(val)

	return c.Next()
}

func (c *JqEachCursor) Next() error {
	v, ok := c.iter.Next()

	if !ok {
		c.rowid = -1
		return nil
	}

	if err, ok := v.(error); ok {
		return fmt.Errorf("error creating result: %w", err)
	}

	c.rowid += 1
	c.value = &v
	return nil
}

func (c *JqEachCursor) Column(ctx *sqlite.VirtualTableContext, i int) error {
	if c.value == nil {
		ctx.ResultNull()
		return nil
	}
	formatResult(ctx, *c.value)
	return nil
}

func (c *JqEachCursor) Eof() bool {
	return c.rowid == -1
}

func (c *JqEachCursor) Close() error { return nil }
