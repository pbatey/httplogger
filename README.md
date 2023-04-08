# httplogger

HTTP request logger middleware for golang.

Heavily inspired by Express.js [morgan](https://github.com/expressjs/morgan).


## API

```
import (
  "github.com/pbatey/httplogger"
)
```

### httplogger.New(next *http.Handler, format string)

Create a new httplogger middleware function using the given format. The format argument may be a string of a predefined names (see below for the names), or a format string of tokens.

### httplogger.New(next *http.Handler, format string)

#### using a predefined format string

```
httplogger.New(next, "tiny")
```

#### using a format of predefined tokens

```
httplogger.New(next, ":method :url :status :res[content-length] - :response-time ms")
```

## Predefined Formats

There are various pre-defined formats provided:

### combined
Standard Apache combined log output.

`:remote-addr - :remote-user [:date[clf]] ":method :url HTTP/:http-version" :status :res[content-length] ":referrer" ":user-agent"`

### common
Standard Apache common log output.

`:remote-addr - :remote-user [:date[clf]] ":method :url HTTP/:http-version" :status :res[content-length]`

### dev
Concise output colored by response status for development use.

`\x1b[0m:method :url :status[clr] :response-time ms - :res[content-length]`

The `\x1b[0m` character ensures color is reset at the beginning of the line.

### short
Shorter than default, also including response time.

`:remote-addr :remote-user :method :url HTTP/:http-version :status :res[content-length] - :response-time ms`

### tiny
The minimal output.

`:method :url :status :res[content-length] - :response-time ms`

## Tokens

### Creating new tokens

To define a token, simply invoke morgan.token() with the name and a callback function. This callback function is expected to return a string value. The value returned is then available as “:type” in this case:

```
httplogger.AddToken('type', func (res http.ResponseWriter, req *http.Request) {
	return req.Header().Get('content-type')
})
```
Calling httplogger.AddToken() using the same name as an existing token will overwrite that token definition.

### :date[format]
The current date and time in UTC. The available formats are:

- `clf` for the common log format ("10/Oct/2000:13:55:36 +0000")
- `iso` for the common ISO 8601 date time format (2000-10-10T13:55:36.000Z)
- `web` for the common RFC 1123 date time format (Tue, 10 Oct 2000 13:55:36 GMT)

If no format is given, then the default is web.

### :http-version
The HTTP version of the request.

### :method
The HTTP method of the request.

### :referrer
The Referrer header of the request. This will use the standard mis-spelled Referer header if exists, otherwise Referrer.

### :remote-addr
The remote address of the request. This will use req.ip, otherwise the standard req.connection.remoteAddress value (socket address).

### :remote-user
The user authenticated as part of Basic auth for the request.

### :req[header]
The given header of the request. If the header is not present, the value will be displayed as "-" in the log.

### :res[header]
The given header of the response. If the header is not present, the value will be displayed as "-" in the log.

If the Content-Length header is missing, the value reported is the number of bytes written.

### :response-time[digits]
The time between the request coming into morgan and when the response headers are written, in milliseconds.

The digits argument is a number that specifies the number of digits to include on the number, defaulting to 3, which provides microsecond precision.

### :status
The status code of the response.

If the request/response cycle completes before a response was sent to the client (for example, the TCP socket closed prematurely by a client aborting the request), then the status will be empty (displayed as "-" in the log).

### :status[clr]
The status code of the response.

The status will be colored. Green for success codes, red for server error codes, yellow for client error codes, cyan for redirection codes, and uncolored for information codes. Use `\1xb[0m` at the beginning of the format to ensure lines start with no color.

### :url
The URL of the request. This will use req.originalUrl if exists, otherwise req.url.

### :user-agent
The contents of the User-Agent header of the request.


## Examples

### A simple file server
```
package main

import (
  "net/http"
  "github.com/pbatey/httplogger"
)

func main() {
	mux := http.NewServeMux()
	hl := httplogger.New(mux, "dev")

	mux.Handle("/", http.FileServer(http.Dir("./public")))

	log.Print("Listening on :3000...")
	err := http.ListenAndServe(":3000", hl)
	log.Fatal(err)
}	
```


## License

[MIT](./LICENSE)

## Acknoledgments

This project is heavily inspired by Express.js [morgan](https://github.com/expressjs/morgan).

_morgan_ is (c) 2014 Jonathan Ong and (c) 2014-2017 Douglas Christopher Wilson. Huge thanks
to them! I sorely missed having _morgan_ in golang so I wrote _httplogger_.

Although this code is not a port and it's source-code is quite different, much of the contents
of this README.md is taken directly from theirs.

I'm a huge fan of _Dexter_ too! I think _Dexter: New Blood_ (2021) made-up, at least a little,
for the end of the original series. I almost named this project _harrison_ 😀.
