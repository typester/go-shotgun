# go-shotgun

Lazy application loader for http application.

## Introduction

There is already many **app reloader** implementations, but most of these are watching **file-system changes**.
I usually use **auto save feature** on Emacs (with [auto-save-buffers-enhanced](https://github.com/kentaro/auto-save-buffers-enhanced)), so these reloader doesn't work expectedly in my environment.

This loader is:

* Act as a HTTP proxy.
* Launch backend application after a HTTP request.
* If some files is changed, re-launch application when **next HTTP request comes**.

## Usage

```
go-shotgun [options] command...
```

Available options are:

* `timeout`: timeout second for waiting application launching (Default 10)
* `map`: port mapping for application (Default: 3000:5000, this means shotgun listens 3000 port and expects `command` uses 5000 port)
* `path`: path for watching filesystem changes

### Examples

Listen 3000 port and auto-reaload app.go (that listen 5000 port)

```
go-shotgun go run app.go
```

Listen 5000 port and auto-reload app.go that listen 5001 port.

```
go-shotgun -map 5000:5001 go run app.go -port 5001
```

This application acts as HTTP-proxy style, so target application don't have to be created by Golang.

```
go-shotgun plackup Hello.psgi
```

Above works fine with plackup command for Perl web application.

## TODO

- [ ] Better console logs
- [ ] If application failed to launch, or died unexpectedly, show these info to browser.
- [ ] Terminate the app by clerner way
- [ ] Spesify watching target files by (regexp?) rules
- [ ] Tests

## Author

Daisuke Murase (typester)





