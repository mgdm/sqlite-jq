# sqlite-jq
An extension for SQLite that adds functions for querying JSON data using JQ syntax.

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

On macOS, run `make`, then you can load the resulting extension into `sqlite3` using `.load sqlite_jq.so`.

On Linux, run `make` to build, though you will then have to place the extension somewhere on `LD_LIBRARY_PATH`. Alternatively, for testing, you can set this directly:

```shell
export LD_LIBRARY_PATH=$PWD:LD_LIBRARY_PATH
```

I would not advise doing this permanently. Then, you can load the resulting extension with `.load sqlite_jq`.

## Things to be aware of

This is, at present, an interesting hack with no tests. I intend to fix this. Notably, I haven't tested the table-valued function with constraints much.

This uses the [gojq](https://github.com/itchyny/gojq) implementation of `jq` by [itchyny](https://github.com/itchyny), which has [some differences](https://github.com/itchyny/gojq#difference-to-jq) from the canonical implementation but is easy to integrate with.

The [sqlite bindings](https://github.com/riyaz-ali/sqlite) in use are by [Riyaz Ali](https://github.com/riyaz-ali).
