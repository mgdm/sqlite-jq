# sqlite-jq
An extension for SQLite that adds functions for querying JSON data using JQ syntax.

## Why?

SQLite does have [JSON functions and operators](https://www.sqlite.org/json1.html) already but I have occasionally done a bit of data mangling in `jq` prior to loading it into SQLite, and sometimes I've wanted to be able to do all of that in one place. `jq`'s language is a bit more flexible than the built-in JSON functions in SQLite for some purposes, particularly when iterating over large deeply-nested objects.

## The `jq()` function

This will return the result of the specified JQ expression run against the supplied JSON.

```sql
select jq('{"a": "xyz"}', '.a');
-- returns "xyz"
```

If there is a single scalar result, it will be returned as the corresponding type. Integers are returned as integers, floats as floats, booleans as ints with value 0 or 1, etc.

If the result is a JSON array or object, those will be returned encoded as JSON.

If there are multiple results they are returned as a JSON array.

## The `jq_each()` table-valued function

This function returns a result set as a virtual table. Each row of the result will be encoded in the same way as above.

```sql
select * from jq_each('{"hello": "world"}', '.hello');
-- returns "world"


-- see test_table.sql for the input
select * from test, jq_each(test.raw, '.[].repo.name');
-- returns two rows, both 'mgdm/htmlq'
```

## Using

### Standalone binary

The easiest way to use sqlite-jq is the self-contained `sqlite3-jq` binary, which is a drop-in replacement for the standard `sqlite3` CLI. The `jq()` and `jq_each()` functions are always available — no `.load` step required.

You'll need the SQLite amalgamation downloaded first (one-time setup):

```shell
make fetch-sqlite
make sqlite3-jq
```

Then use it exactly like the standard `sqlite3` shell:

```shell
./sqlite3-jq :memory: "SELECT jq('{\"a\":1}', '.a')"
# 1

./sqlite3-jq :memory: "SELECT value FROM jq_each('[1,2,3]', '.[]')"
# 1
# 2
# 3

./sqlite3-jq mydata.db
```

### Dynamic library

On macOS, run `make`, then load the extension into `sqlite3`:

```shell
make
sqlite3 mydata.db
sqlite> .load sqlite_jq.dylib
```

On Linux, run `make` to build. You will need to place the extension somewhere on `LD_LIBRARY_PATH`, or for a quick test:

```shell
export LD_LIBRARY_PATH=$PWD:$LD_LIBRARY_PATH
sqlite3 mydata.db ".load sqlite_jq"
```

## Known issues

When using the dynamic library with the standard `sqlite3` shell, if you open a new database within the same session you'll need to re-load the extension. This doesn't affect the standalone `sqlite3-jq` binary, which registers the extension for every connection automatically.

You can't currently write a query like this:

```sql
select * from jq_each(raw_data.raw, '.things[]');
```

instead you must write it as follows:

```sql
select * from raw_data, jq_each(raw, '.things[]');
```

## Things to be aware of

This is an interesting hack. The table-valued function with non-trivial constraints hasn't been tested much.

This uses the [gojq](https://github.com/itchyny/gojq) implementation of `jq` by [itchyny](https://github.com/itchyny), which has [some differences](https://github.com/itchyny/gojq#difference-to-jq) from the canonical implementation but is easy to integrate with.

The [sqlite bindings](https://github.com/riyaz-ali/sqlite) in use are by [Riyaz Ali](https://github.com/riyaz-ali).
