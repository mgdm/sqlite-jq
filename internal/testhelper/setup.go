// Package testhelper wires sqlite3_extension_init into sqlite3_auto_extension
// so that every connection opened by mattn/go-sqlite3 gets the jq functions.
// This replicates the setup from go.riyazali.net/sqlite/internal/testing/sqlite.
package testhelper

// #cgo CFLAGS: -DSQLITE_CORE -Wno-deprecated-declarations
// #include <sqlite3.h>
// extern int sqlite3_extension_init(sqlite3*, char**, const sqlite3_api_routines*);
import "C"

func init() {
	C.sqlite3_auto_extension((*[0]byte)(C.sqlite3_extension_init))
}
