.load sqlite_jq
.echo on

select jq('{"a": "xyz"}', '.a');
select typeof(jq('{"a": "xyz"}', '.a'));

select jq('{"a": 0.5}', '.a');
select typeof(jq('{"a": 0.5}', '.a'));

select jq('{"a": "xyz"}', '.b');
select typeof(jq('{"a": "xyz"}', '.b'));

select jq('[{"a": "xyz"}, {"a": "123"}]', '.[].a');
select typeof(jq('[{"a": "xyz"}, {"a": "123"}]', '.[].a'));

select jq('["foo", "bar", "baz"]', '@csv');
select typeof(jq('["foo", "bar", "baz"]', '@csv'));

select jq('["foo", "bar", "baz"]', '@csv');
select typeof(jq('["foo", "bar", "baz"]', '@csv'));

select jq('{"a": {"hello": "world"}}', '.a');
select typeof(jq('{"a": {"hello": "world"}}', '.a'));

select json_extract(jq('{"a": {"hello": "world"}}', '.a'), '$.hello');
select typeof(json_extract(jq('{"a": {"hello": "world"}}', '.a'), '$.hello'));

select jq('{"a": ["1", "2", "3"]}', '.a');
select typeof(jq('{"a": ["1", "2", "3"]}', '.a'));

select jq('{"a": [1, 2, 3]}', '.a');
select typeof(jq('{"a": [1, 2, 3]}', '.a'));

select jq('[{"a": [1, 2, 3]}, {"a": 1}]', '.[].a');
select typeof(jq('[{"a": [1, 2, 3]}, {"a": 1}]', '.[].a'));

select jq('[]', 'this is not a query');

select jq('this is not json', 'this is not a query');

select jq(null, '.a');
select jq('', '.a');
select jq('[]', null);
