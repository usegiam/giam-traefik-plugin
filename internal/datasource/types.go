package datasource

type Datasource string

var (
	Loki       Datasource = "loki"
	Prometheus Datasource = "prometheus"
)
