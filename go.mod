module github.com/azhai/gopaper

go 1.27

require (
	github.com/VictoriaMetrics/fastcache v1.13.3
	github.com/azhai/gobus v0.2.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/labstack/echo/v5 v5.4.2
	github.com/pelletier/go-toml/v2 v2.4.3
	github.com/yuin/goldmark v1.8.6
	golang.org/x/crypto v0.58.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/time v0.16.0 // indirect
)

replace github.com/azhai/gobus => ../gobus
