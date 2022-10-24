.load sqlite_jq

create table test(
      raw TEXT NOT NULL
);

insert into test values ('[
  {
    "id": "24583862139",
    "type": "PushEvent",
    "actor": {
      "id": 71893,
      "login": "mgdm",
      "display_login": "mgdm",
      "gravatar_id": "",
      "url": "https://api.github.com/users/mgdm",
      "avatar_url": "https://avatars.githubusercontent.com/u/71893?"
    },
    "repo": {
      "id": 185476675,
      "name": "mgdm/htmlq",
      "url": "https://api.github.com/repos/mgdm/htmlq"
    },
    "payload": {
      "push_id": 11320965987,
      "size": 1,
      "distinct_size": 1,
      "ref": "refs/heads/master",
      "head": "739cd363543cd5c36a2d7bcbbb3ab7e811205611",
      "before": "1f5fa50722436df15d57e8627e32b68a6dc8c927",
      "commits": [
        {
          "sha": "739cd363543cd5c36a2d7bcbbb3ab7e811205611",
          "author": {
            "email": "michael@mgdm.net",
            "name": "Michael Maclean"
          },
          "message": "Add flake.nix",
          "distinct": true,
          "url": "https://api.github.com/repos/mgdm/htmlq/commits/739cd363543cd5c36a2d7bcbbb3ab7e811205611"
        }
      ]
    },
    "public": true,
    "created_at": "2022-10-13T17:52:21Z"
  },
  {
    "id": "24583836616",
    "type": "PushEvent",
    "actor": {
      "id": 71893,
      "login": "mgdm",
      "display_login": "mgdm",
      "gravatar_id": "",
      "url": "https://api.github.com/users/mgdm",
      "avatar_url": "https://avatars.githubusercontent.com/u/71893?"
    },
    "repo": {
      "id": 185476675,
      "name": "mgdm/htmlq",
      "url": "https://api.github.com/repos/mgdm/htmlq"
    },
    "payload": {
      "push_id": 11320953650,
      "size": 1,
      "distinct_size": 1,
      "ref": "refs/heads/master",
      "head": "1f5fa50722436df15d57e8627e32b68a6dc8c927",
      "before": "103bb2157fba78218e2679ce16365a769de12ccf",
      "commits": [
        {
          "sha": "1f5fa50722436df15d57e8627e32b68a6dc8c927",
          "author": {
            "email": "michael.maclean@bbc.co.uk",
            "name": "Michael Maclean"
          },
          "message": "Add flake.nix",
          "distinct": true,
          "url": "https://api.github.com/repos/mgdm/htmlq/commits/1f5fa50722436df15d57e8627e32b68a6dc8c927"
        }
      ]
    },
    "public": true,
    "created_at": "2022-10-13T17:51:04Z"
  }
]');

select * from test, jq_each(test.raw, '.[].repo.name');

.exit
