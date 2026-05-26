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
	var args int
	var jsonArgv, queryArgv int

	output := &sqlite.IndexInfoOutput{
		ConstraintUsage: make([]*sqlite.ConstraintUsage, len(input.Constraints)),
		EstimatedCost:   1000,
	}

	for j, con := range input.Constraints {
		// Only consume EQ constraints on the two hidden input columns.
		if con.ColumnIndex < 1 || con.Op != sqlite.INDEX_CONSTRAINT_EQ {
			continue
		}
		if !con.Usable {
			return nil, sqlite.SQLITE_CONSTRAINT
		}
		args++
		output.ConstraintUsage[j] = &sqlite.ConstraintUsage{ArgvIndex: args, Omit: true}
		switch con.ColumnIndex {
		case 1:
			jsonArgv = args
		case 2:
			queryArgv = args
		}
	}

	// Encode which argv slot holds json (low nibble) and query (high nibble).
	// A zero nibble means that parameter was not provided.
	output.IndexNumber = jsonArgv | (queryArgv << 4)
	return output, nil
}

func (c *JqEachCursor) Rowid() (int64, error) {
	return c.rowid, nil
}

func (c *JqEachCursor) Filter(idxNum int, _ string, values ...sqlite.Value) error {
	jsonArgv := idxNum & 0xf
	queryArgv := (idxNum >> 4) & 0xf

	if jsonArgv == 0 || queryArgv == 0 {
		c.rowid = -1
		return sqlite.SQLITE_OK
	}

	jsonVal := values[jsonArgv-1]
	queryVal := values[queryArgv-1]

	if jsonVal.Type() == sqlite.SQLITE_NULL || queryVal.Type() == sqlite.SQLITE_NULL {
		c.rowid = -1
		return sqlite.SQLITE_OK
	}

	var val interface{}
	dec := json.NewDecoder(bytes.NewReader(jsonVal.Blob()))
	dec.UseNumber()
	if err := dec.Decode(&val); err != nil {
		return fmt.Errorf("error parsing JSON data: %w", err)
	}

	code, err := compileQuery(queryVal.Text())
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
