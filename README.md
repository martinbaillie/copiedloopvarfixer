# copiedloopvarfixer

> Fixes your Go code's copied loop vars

[Go 1.22](https://go.dev/blog/go1.22) resolves the long-standing loop variable gotcha.

This is a small program you can use to recursively find and remove copied loop variables in your
codebase.

## Usage

```shell
go get github.com/martinbaillie/copiedloopvarfixer
go run github.com/martinbaillie/copiedloopvarfixer <dir_to_walk>
```

## Disclaimer
Check the files changed by this program and ensure your code still tests and runs correctly. I take
zero responsibility for any destructive changes this may cause to your runtime.
