ifeq ($(shell uname -s),Darwin)
LIBRARY_EXTENSION=dylib
else ifeq ($(OS),Windows_NT)
LIBRARY_EXTENSION=dll
else
LIBRARY_EXTENSION=so
endif

build:
	go build -buildmode=c-shared -o sqlite_jq.$(LIBRARY_EXTENSION) ./*.go

test-sql: build
	sqlite3 :memory: < test.sql

test-table: build
	sqlite3 :memory: < test_table.sql

test: build test-sql test-table

clean:
	rm -f *.$(LIBRARY_EXTENSION)

goimports:
	goimports -w *.go
