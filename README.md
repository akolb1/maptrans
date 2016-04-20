# maptrans

maptrans is a Go library for translating maps into other maps. 

The library is useful when dealing with JSON-based APIs which often use 
map[string]interface{} as a data structure decoded from JSON. This often 
should be re-packaged and sent to other JSON APIs. The maptrans library 
provides a descriptive way to specify such map translations.

## Installation

Standard `go get`:

```
$ go get github.com/akolb1/maptrans

## Usage & Example

For usage and examples see the [Godoc](http://godoc.org/github.com/akolb1/maptrans).
