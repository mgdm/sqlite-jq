package main

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/mgdm/sqlite-jq/internal/testhelper"
)

func openDB(t testing.TB) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func queryOne(t testing.TB, db *sql.DB, query string, args ...interface{}) *sql.Row {
	t.Helper()
	return db.QueryRow(query, args...)
}

// jq() scalar function tests

func TestJqString(t *testing.T) {
	db := openDB(t)
	var got string
	if err := queryOne(t, db, `SELECT jq('{"a":"xyz"}', '.a')`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != "xyz" {
		t.Errorf("got %q, want %q", got, "xyz")
	}
}

func TestJqInt(t *testing.T) {
	db := openDB(t)
	var got int64
	if err := queryOne(t, db, `SELECT jq('{"n":42}', '.n')`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestJqFloat(t *testing.T) {
	db := openDB(t)
	var got float64
	if err := queryOne(t, db, `SELECT jq('{"x":1.5}', '.x')`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 1.5 {
		t.Errorf("got %f, want 1.5", got)
	}
}

func TestJqBoolTrue(t *testing.T) {
	db := openDB(t)
	var got int64
	if err := queryOne(t, db, `SELECT jq('{"b":true}', '.b')`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Errorf("got %d, want 1", got)
	}
}

func TestJqBoolFalse(t *testing.T) {
	db := openDB(t)
	var got int64
	if err := queryOne(t, db, `SELECT jq('{"b":false}', '.b')`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestJqObject(t *testing.T) {
	db := openDB(t)
	var got string
	if err := queryOne(t, db, `SELECT jq('{"a":{"k":"v"}}', '.a')`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != `{"k":"v"}` {
		t.Errorf("got %q, want %q", got, `{"k":"v"}`)
	}
}

func TestJqArray(t *testing.T) {
	db := openDB(t)
	var got string
	if err := queryOne(t, db, `SELECT jq('{"a":[1,2,3]}', '.a')`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != `[1,2,3]` {
		t.Errorf("got %q, want %q", got, `[1,2,3]`)
	}
}

func TestJqMultipleResults(t *testing.T) {
	db := openDB(t)
	var got string
	if err := queryOne(t, db, `SELECT jq('[{"a":"x"},{"a":"y"}]', '.[].a')`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != `["x","y"]` {
		t.Errorf("got %q, want %q", got, `["x","y"]`)
	}
}

func TestJqNullOutput(t *testing.T) {
	db := openDB(t)
	var got interface{}
	if err := queryOne(t, db, `SELECT jq('{"a":null}', '.a')`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestJqMissingKey(t *testing.T) {
	db := openDB(t)
	var got interface{}
	if err := queryOne(t, db, `SELECT jq('{"a":"x"}', '.b')`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestJqInvalidJSON(t *testing.T) {
	db := openDB(t)
	var got interface{}
	err := queryOne(t, db, `SELECT jq('not json', '.a')`).Scan(&got)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestJqInvalidQuery(t *testing.T) {
	db := openDB(t)
	var got interface{}
	err := queryOne(t, db, `SELECT jq('[]', 'this is not valid jq')`).Scan(&got)
	if err == nil {
		t.Error("expected error for invalid jq query, got nil")
	}
}

func TestJqNullInputJSON(t *testing.T) {
	db := openDB(t)
	var got interface{}
	// NULL blob is treated as empty → parse error or NULL result
	if err := queryOne(t, db, `SELECT jq(NULL, '.a')`).Scan(&got); err != nil {
		// error is acceptable
		return
	}
	// NULL result is also acceptable
}

// jq_each() table-valued function tests

func TestJqEachScalars(t *testing.T) {
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each('[1,2,3]', '.[]')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var got []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []int64{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestJqEachStrings(t *testing.T) {
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each('["a","b","c"]', '.[]')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestJqEachNullOutput(t *testing.T) {
	// Regression: jq null output must produce SQL NULL, not the blob "null"
	db := openDB(t)
	rows, err := db.Query(`SELECT value, typeof(value) FROM jq_each('[null]', '.[]')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("expected one row")
	}
	var val interface{}
	var typ string
	if err := rows.Scan(&val, &typ); err != nil {
		t.Fatal(err)
	}
	if val != nil {
		t.Errorf("value: got %v, want nil", val)
	}
	if typ != "null" {
		t.Errorf("typeof: got %q, want %q", typ, "null")
	}
}

func TestJqEachNullInput(t *testing.T) {
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each(NULL, '.[]')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if count != 0 {
		t.Errorf("got %d rows for NULL input, want 0", count)
	}
}

func TestJqEachObjects(t *testing.T) {
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each('[{"k":1},{"k":2}]', '.[]')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("got %d rows, want 2", count)
	}
}

func TestJqEachEmpty(t *testing.T) {
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each('[]', '.[]')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	if count != 0 {
		t.Errorf("got %d rows for empty array, want 0", count)
	}
}

func TestJqEachInvalidJSON(t *testing.T) {
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each('not json', '.[]')`)
	if err != nil {
		// Error at query time is acceptable.
		return
	}
	defer rows.Close()
	// Error may also surface when iterating rows.
	for rows.Next() {
	}
	if err := rows.Err(); err == nil {
		t.Error("expected error for invalid JSON input, got nil")
	}
}

func TestJqEachInvalidQuery(t *testing.T) {
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each('[1,2,3]', '???invalid')`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
	}
	if err := rows.Err(); err == nil {
		t.Error("expected error for invalid jq query, got nil")
	}
}

func TestJqEachMixedTypes(t *testing.T) {
	db := openDB(t)
	rows, err := db.Query(`SELECT value, typeof(value) FROM jq_each('[1,"a",null,true]', '.[]')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	type row struct {
		val interface{}
		typ string
	}
	var got []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.val, &r.typ); err != nil {
			t.Fatal(err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	if len(got) != 4 {
		t.Fatalf("got %d rows, want 4", len(got))
	}
	if got[0].typ != "integer" {
		t.Errorf("row 0 type: got %q, want %q", got[0].typ, "integer")
	}
	if got[1].typ != "text" {
		t.Errorf("row 1 type: got %q, want %q", got[1].typ, "text")
	}
	if got[2].typ != "null" {
		t.Errorf("row 2 type: got %q, want %q", got[2].typ, "null")
	}
	if got[3].typ != "integer" {
		t.Errorf("row 3 type (bool): got %q, want %q", got[3].typ, "integer")
	}
}

func TestJqEachSQLWhere(t *testing.T) {
	// SQLite post-filters on the value column; jq_each returns all 3 rows,
	// SQLite discards the one that doesn't satisfy value > 1.
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each('[1,2,3]', '.[]') WHERE value > 1`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var got []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if len(got) != 2 {
		t.Fatalf("got %v, want [2 3]", got)
	}
	if got[0] != 2 || got[1] != 3 {
		t.Errorf("got %v, want [2 3]", got)
	}
}

func TestJqEachDeepExtraction(t *testing.T) {
	db := openDB(t)
	var got int64
	err := db.QueryRow(`SELECT value FROM jq_each('[{"a":{"b":42}}]', '.[].a.b')`).Scan(&got)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestJqEachMultiExpr(t *testing.T) {
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each('{"x":1,"y":2}', '.x, .y')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var got []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if len(got) != 2 {
		t.Fatalf("got %v, want [1 2]", got)
	}
	if got[0] != 1 || got[1] != 2 {
		t.Errorf("got %v, want [1 2]", got)
	}
}

func TestJqEachJoin(t *testing.T) {
	db := openDB(t)
	if _, err := db.Exec(`CREATE TABLE t(raw TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO t VALUES ('[{"name":"alice"}]'), ('[{"name":"bob"}]')`); err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query(`SELECT value FROM t, jq_each(t.raw, '.[].name') ORDER BY value`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	want := []string{"alice", "bob"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestJqEachRowOrder(t *testing.T) {
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each('["c","a","b"]', '.[]')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var got []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	want := []string{"c", "a", "b"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("position %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

// Tests that required the BestIndex/Filter code fix:

func TestJqEachMissingQuery(t *testing.T) {
	// Calling jq_each with only the json argument (no query) must return 0 rows,
	// not panic with an index-out-of-range accessing values[1].
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each('[1,2,3]')`)
	if err != nil {
		// Error at query-plan time is also acceptable.
		return
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	if err := rows.Err(); err != nil {
		// Runtime error is acceptable; what's not acceptable is a panic.
		return
	}
	if count != 0 {
		t.Errorf("got %d rows with missing query arg, want 0", count)
	}
}

func TestJqEachNullQuery(t *testing.T) {
	db := openDB(t)
	rows, err := db.Query(`SELECT value FROM jq_each('[1,2,3]', NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("got %d rows for NULL query, want 0", count)
	}
}
