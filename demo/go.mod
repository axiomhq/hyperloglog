module github.com/axiomhq/hyperloglog/demo

go 1.12

require (
	github.com/axiomhq/hyperloglog v0.0.0-00010101000000-000000000000
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/influxdata/influxdb v1.7.6
)

replace github.com/axiomhq/hyperloglog => ../
