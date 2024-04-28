﻿# simple-data-server


## https
  - `https://sds.merith.xyz/api/<object>/<table>/<key>`
    - `GET` 
      - returns table as json, if key is defined, returns key as text
    - `POST`
      - sets `object/table/key` to value
    - `DELETE`
      - removes `key` from table
## BasicAuth
  - Supply BasicAuth Credentials to use an seperate object/table tree, endpoints remain unchanged
## websocket
  - endpoint: `api/<object>/<table>/ws`
  - onConnect: returns `uuid: uuid`
    - uuid is generated from your basicAuth login
    - if you did not supply one, `default` is used
  - `set key value`
    - sets `object/table/key` to value
    - returns `set: key` over websocket
  - `get key`
    - returns `value`
  - `del key`
    - removes key from table
