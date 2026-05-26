ifeq ($(shell uname -s),Darwin)
LIBRARY_EXTENSION=dylib
else ifeq ($(OS),Windows_NT)
LIBRARY_EXTENSION=dll
else
LIBRARY_EXTENSION=so
endif

SQLITE_VERSION_NUM = 3530000
SQLITE_YEAR        = 2026
SQLITE_URL         = https://www.sqlite.org/$(SQLITE_YEAR)/sqlite-amalgamation-$(SQLITE_VERSION_NUM).zip
GO_LDFLAGS         = $(shell go env CGO_LDFLAGS)

build:
	go build -buildmode=c-shared -o sqlite_jq.$(LIBRARY_EXTENSION) ./*.go

fetch-sqlite: sqlite/shell.c

sqlite/shell.c:
	mkdir -p sqlite
	curl -Lf -o sqlite/amalgamation.zip $(SQLITE_URL)
	cd sqlite && unzip -j amalgamation.zip '*/shell.c' '*/sqlite3.c' '*/sqlite3.h' '*/sqlite3ext.h'
	rm sqlite/amalgamation.zip

sqlite_jq.a: $(wildcard *.go)
	go build -buildmode=c-archive -o sqlite_jq.a ./*.go

sqlite3-jq: sqlite_jq.a sqlite/shell.c standalone/init_shim.c
	$(CC) -o sqlite3-jq \
	    -DSQLITE_SHELL_INIT_PROC=register_jq_extension \
	    sqlite/sqlite3.c \
	    sqlite/shell.c \
	    standalone/init_shim.c \
	    sqlite_jq.a \
	    $(GO_LDFLAGS) \
	    -lpthread -ldl \
	    -Wno-deprecated-declarations

test-sql: build
	sqlite3 :memory: < test.sql; true

test-table: build
	sqlite3 :memory: < test_table.sql; true

test: build test-sql test-table test-go

test-go:
	go test ./...

clean:
	rm -f *.$(LIBRARY_EXTENSION) *.h *.a sqlite3-jq
	rm -rf sqlite/

goimports:
	goimports -w *.go
