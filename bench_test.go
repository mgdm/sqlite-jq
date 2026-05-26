package main

import (
	"testing"

	"github.com/itchyny/gojq"
)

// smallJSON / largeJSON are reused across benchmarks to isolate the
// parse+compile cost from JSON unmarshal cost.
const smallJSON = `{"repo":{"name":"mgdm/htmlq"},"type":"PushEvent"}`

const largeJSON = `[{"id":"24583862139","type":"PushEvent","actor":{"id":71893,"login":"mgdm","display_login":"mgdm","gravatar_id":"","url":"https://api.github.com/users/mgdm","avatar_url":"https://avatars.githubusercontent.com/u/71893?"},"repo":{"id":185476675,"name":"mgdm/htmlq","url":"https://api.github.com/repos/mgdm/htmlq"},"payload":{"push_id":11320965987,"size":1,"distinct_size":1,"ref":"refs/heads/master","head":"739cd363543cd5c36a2d7bcbbb3ab7e811205611","before":"1f5fa50722436df15d57e8627e32b68a6dc8c927","commits":[{"sha":"739cd363543cd5c36a2d7bcbbb3ab7e811205611","author":{"email":"michael@mgdm.net","name":"Michael Maclean"},"message":"Add flake.nix","distinct":true}]},"public":true,"created_at":"2022-10-13T17:52:21Z"},{"id":"24583836616","type":"PushEvent","actor":{"id":71893,"login":"mgdm","display_login":"mgdm","gravatar_id":"","url":"https://api.github.com/users/mgdm","avatar_url":"https://avatars.githubusercontent.com/u/71893?"},"repo":{"id":185476675,"name":"mgdm/htmlq","url":"https://api.github.com/repos/mgdm/htmlq"},"payload":{"push_id":11320953650,"size":1,"distinct_size":1,"ref":"refs/heads/master","head":"1f5fa50722436df15d57e8627e32b68a6dc8c927","before":"103bb2157fba78218e2679ce16365a769de12ccf","commits":[{"sha":"1f5fa50722436df15d57e8627e32b68a6dc8c927","author":{"email":"michael.maclean@bbc.co.uk","name":"Michael Maclean"},"message":"Add flake.nix","distinct":true}]},"public":true,"created_at":"2022-10-13T17:51:04Z"}]`

// --- compile step in isolation ---

// BenchmarkCompileHot measures a cache hit: parse+compile already done.
func BenchmarkCompileHot(b *testing.B) {
	const q = ".repo.name"
	compileQuery(q) // warm cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileQuery(q)
	}
}

// BenchmarkCompileCold measures what happened before the cache: parse+compile
// on every call.
func BenchmarkCompileCold(b *testing.B) {
	const q = ".repo.name"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsed, _ := gojq.Parse(q)
		gojq.Compile(parsed)
	}
}

// --- end-to-end jq() via SQL ---

func BenchmarkJqScalarHot(b *testing.B) {
	db := openDB(b)
	var got string
	db.QueryRow(`SELECT jq(?, '.repo.name')`, smallJSON).Scan(&got) // warm cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.QueryRow(`SELECT jq(?, '.repo.name')`, smallJSON).Scan(&got)
	}
}

func BenchmarkJqScalarCold(b *testing.B) {
	db := openDB(b)
	var got string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queryCache.Delete(".repo.name")
		db.QueryRow(`SELECT jq(?, '.repo.name')`, smallJSON).Scan(&got)
	}
}

func BenchmarkJqLargeJSONHot(b *testing.B) {
	db := openDB(b)
	var got string
	db.QueryRow(`SELECT jq(?, '.[].repo.name')`, largeJSON).Scan(&got) // warm cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.QueryRow(`SELECT jq(?, '.[].repo.name')`, largeJSON).Scan(&got)
	}
}

func BenchmarkJqLargeJSONCold(b *testing.B) {
	db := openDB(b)
	var got string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queryCache.Delete(".[].repo.name")
		db.QueryRow(`SELECT jq(?, '.[].repo.name')`, largeJSON).Scan(&got)
	}
}

// --- end-to-end jq_each() via SQL ---

func BenchmarkJqEachHot(b *testing.B) {
	db := openDB(b)
	db.QueryRow(`SELECT value FROM jq_each(?, '.[].repo.name') LIMIT 1`, largeJSON).Scan(nil) // warm cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Query(`SELECT value FROM jq_each(?, '.[].repo.name')`, largeJSON)
		for rows.Next() {
			var v string
			rows.Scan(&v)
		}
		rows.Close()
	}
}

func BenchmarkJqEachCold(b *testing.B) {
	db := openDB(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queryCache.Delete(".[].repo.name")
		rows, _ := db.Query(`SELECT value FROM jq_each(?, '.[].repo.name')`, largeJSON)
		for rows.Next() {
			var v string
			rows.Scan(&v)
		}
		rows.Close()
	}
}
