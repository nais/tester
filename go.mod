module github.com/nais/tester

go 1.24

tool (
	golang.org/x/tools/cmd/deadcode
	golang.org/x/vuln/cmd/govulncheck
	honnef.co/go/tools/cmd/staticcheck
	mvdan.cc/gofumpt
)

require (
	github.com/fsnotify/fsnotify v1.8.0
	github.com/google/go-cmp v0.7.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/mitchellh/mapstructure v1.5.0
	github.com/yuin/gopher-lua v1.1.1
	golang.org/x/sync v0.11.0
)

require (
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/exp/typeparams v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/telemetry v0.0.0-20240522233618-39ace7a40ae7 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
	golang.org/x/vuln v1.1.4 // indirect
	honnef.co/go/tools v0.6.0 // indirect
	mvdan.cc/gofumpt v0.7.0 // indirect
)
