module github.com/vedranvuk/config

go 1.14

require (
	github.com/vedranvuk/errorex v0.3.2
	github.com/vedranvuk/reflectex v0.0.1
	github.com/vedranvuk/typeregistry v0.0.2
)

replace (
	github.com/vedranvuk/errorex => ../errorex
	github.com/vedranvuk/reflectex => ../reflectex
	github.com/vedranvuk/typeregistry => ../typeregistry
)
