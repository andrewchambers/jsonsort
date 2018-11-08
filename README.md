# jsonsort

Sort a stream of line delimited json objects or arrays by a key using coreutils sort.

## Rationale

coreutils sort can handle large amounts of data using temporary files,
most json sorting tools I have seen do not support this, This tool should
be able to sort lists of json objects much larger than what fits in memory.

## Flags

```
Usage of jsonsort:
  -command string
        Sort command binary (usually should be gnu sort). (default "sort")
  -debug
        Debug output to stderr.
  -ignore-case
        Ignore case in key.
  -method string
        For gnusort, valid methods are general-numeric, human-numeric, month, numeric, random or version sort.
  -unique
         Don't print json objects with duplicate keys, unique keys only.

```

## Usage examples

```
$ cat data.jsonl
[{"key":"1kb","b":2}]
[{"key":"1gb","b":1}]

# Sort by first element in array and field 'key' using human numeric sorting (see coreutils manual).
$ jsonsort -method human-numeric 0 key < data.jsonl
[{"key":"1gb","b":1}]
[{"key":"1kb","b":2}]

# Sort nobel laureates by first name.
curl http://api.nobelprize.org/v1/prize.json | jq -c .prizes[].laureates | jsonsort -ignore-case firstname
```

## Usage notes

- jsonsort expects one json object per line.
- jsonsort uses the byte 0x02 as a delimiter for the sort tool.
  A sort key containing this value will be sorted incorrectly.
- The first non flag argument begins the key list.
- Use '--' to force the key list to start.


## Building

Ensure you have a version of 'go' that supports modules, then type 'go build' from
the base dir.

## TODO

- The flags could be better designed, perhaps even a thinner wrapper
  where arbitrary sort flags can be passed through.
- Be compatible with POSIX sort, not just gnu sort.
- Man page.

## Packagers/maintainers

Create an issue to discuss packaging for your platform, we can negotiate command
line stabilization, or perhaps a versioning scheme like 'jsonsort2' etc.

