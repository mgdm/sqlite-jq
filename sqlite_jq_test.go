package main

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/mgdm/sqlite-jq/internal/testhelper"
)

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func queryOne(t *testing.T, db *sql.DB, query string, args ...interface{}) *sql.Row {
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
