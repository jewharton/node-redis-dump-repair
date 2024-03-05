# Repair Tool for node-redis-dump

This tool repairs dump files produced by the [redis-dump](https://www.npmjs.com/package/redis-dump) NPM package which may contain inappropriate line breaks that prevent the file from being loaded into a Redis database.

## Installation

```console
$ git clone https://github.com/jewharton/node-redis-dump-repair
$ cd node-redis-dump-repair
$ go install ./...
```

## Usage

```console
$ redis-dump-repair --help
Repair malformed dumps produced by the redis-dump NPM package.

Usage:
  redis-dump-repair [input-file] [output-file] [flags]

Flags:
  -h, --help   help for redis-dump-repair
```

## Motivation

When dumping a Redis database using node-redis-dump's `redis-dump` command, the resulting file may contain unescaped newline characters. This file is meant to be loaded back into a Redis database by running the following command:

```console
$ cat dump.txt | redis-cli
```

`redis-cli` uses newline characters to terminate Redis commands. Because `redis-dump` doesn't escape newlines, each key or value that contains them will be broken into multiple commands.

```console
$ redis-cli
127.0.0.1:6379> set foo "bar\nbaz"
OK
127.0.0.1:6379> quit
$ redis-dump > dump.txt
$ cat dump.txt
SET     foo 'bar
baz'
```

In the above example, `SET foo 'bar` and `baz'` are considered separate commands by `redis-cli` because of the newline character between them. The expected behavior is that they would remain on a single line as `SET foo "bar\nbaz"`.


