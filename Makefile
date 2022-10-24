build:
	go build -buildmode=c-shared -o sqlite_jq.so ./*.go

test-sql: build
	sqlite3 :memory: < test.sql

test-table: build
	sqlite3 :memory: < test_table.sql

clean:
	rm -f *.so

goimports:
	goimports -w *.go
