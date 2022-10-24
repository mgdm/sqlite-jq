package main

import (
	"encoding/json"
	"fmt"

	"github.com/itchyny/gojq"
	"go.riyazali.net/sqlite"
)

type Jq struct{}

func (m *Jq) Args() int           { return 2 }
func (m *Jq) Deterministic() bool { return true }
func (m *Jq) Apply(ctx *sqlite.Context, values ...sqlite.Value) {
	var val interface{}
	err := json.Unmarshal(values[0].Blob(), &val)

	if err != nil {
		err = fmt.Errorf("error parsing JSON data: %v", err)
		ctx.ResultError(err)
		return
	}

	query, err := gojq.Parse(values[1].Text())

	if err != nil {
		err = fmt.Errorf("error parsing JQ query: %v", err)
		ctx.ResultError(err)
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
			err = fmt.Errorf("error creating result: %v", err)
			ctx.ResultError(err)
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

func formatResult(ctx *sqlite.Context, v interface{}) {
	if v == nil {
		ctx.ResultNull()
		return
	}

	switch v.(type) {
	case bool:
		if v.(bool) {
			ctx.ResultInt(1)
		} else {
			ctx.ResultInt(0)
		}
	case int:
		ctx.ResultInt(v.(int))
	case int64:
		ctx.ResultInt64(v.(int64))
	case float64:
		ctx.ResultFloat(v.(float64))
	case string:
		ctx.ResultText(v.(string))
	default:
		tmp, err := json.Marshal(v)

		if err != nil {
			err = fmt.Errorf("error marshalling result data: %v", err)
			ctx.ResultError(err)
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
