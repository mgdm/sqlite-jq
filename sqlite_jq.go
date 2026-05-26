package main

import (
	"encoding/json"
	"fmt"

	"github.com/itchyny/gojq"
	"go.riyazali.net/sqlite"
)

type resultSetter interface {
	ResultNull()
	ResultInt(int)
	ResultInt64(int64)
	ResultFloat(float64)
	ResultText(string)
	ResultBlob([]byte)
	ResultError(error)
}

type Jq struct{}

func (m *Jq) Args() int           { return 2 }
func (m *Jq) Deterministic() bool { return true }
func (m *Jq) Apply(ctx *sqlite.Context, values ...sqlite.Value) {
	var val interface{}
	err := json.Unmarshal(values[0].Blob(), &val)

	if err != nil {
		ctx.ResultError(fmt.Errorf("error parsing JSON data: %w", err))
		return
	}

	query, err := gojq.Parse(values[1].Text())

	if err != nil {
		ctx.ResultError(fmt.Errorf("error parsing JQ query: %w", err))
		return
	}

	var rows []interface{}

	iter := query.Run(val)
	for {
		v, ok := iter.Next()

		if !ok {
			break
		}

		if err, ok := v.(error); ok {
			ctx.ResultError(fmt.Errorf("error creating result: %w", err))
			return
		}

		rows = append(rows, v)
	}

	switch len(rows) {
	case 0:
		ctx.ResultNull()
	case 1:
		formatResult(ctx, rows[0])
	default:
		formatResult(ctx, rows)
	}
}

func formatResult(ctx resultSetter, v interface{}) {
	if v == nil {
		ctx.ResultNull()
		return
	}

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
			ctx.ResultError(fmt.Errorf("error marshalling result data: %w", err))
			return
		}

		ctx.ResultBlob(tmp)
	}
}

func init() {
	sqlite.Register(func(api *sqlite.ExtensionApi) (sqlite.ErrorCode, error) {
		if err := api.CreateFunction("jq", &Jq{}); err != nil {
			return sqlite.SQLITE_ERROR, err
		}

		if err := api.CreateModule("jq_each", &JqEachModule{}, sqlite.EponymousOnly(true)); err != nil {
			return sqlite.SQLITE_ERROR, err
		}

		return sqlite.SQLITE_OK, nil
	})
}

func main() {}
