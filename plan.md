# structure
- `/api/<dataset>/<unique-id>/<key>`
    - dataset is the project type, this is typically hardocded
    - unique-id is the unique ID of the data being accessed, typically an UUID
    - key, `ws` is the websocket for the dataset, but if you just use the value name, it will return that value
- `/api/<dataset>/<unique-id>` returns json of all of the key pairs
# data
- all data is stored as either ints, bools, or strings.
- data is stored as json files per dataset, per unique-id

# websocket
- commands
    - `set key value`
        - responds with `set key` when backend completes the action
    - `get key`
        - responds with `key value` when backend completes the action


