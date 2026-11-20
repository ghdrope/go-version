# go-version

Go package for version and build info.

It helps you expose version information (version, git commit, build date, Go version and platform) across different Go projects.

---

## Setup

Add the package to your project:

```bash
go get github.com/ghdrope/go-version
```

---

## Usage

### Basic Example

```go
import (
  "fmt"
  "github.com/ghdrope/go-version"
)

func main() {
  fmt.Println(version.String())
}
```

### Short Version Only

```go
fmt.Println(version.Short())
```

### Output

`version.String()`:

```text
Version   : v1.0.0
GitCommit : abc123
BuildDate : 2026-01-01
GoVersion : go1.27.4
Platform  : linux/amd64
```

`version.Short()`:

```text
Version:    v1.0.0
```

Without `ldflags` (e.g. `go run`), the defaults are `development`, `none` and `unknown`.

### Individual Values

The build values are exported variables and can be read directly:

```go
fmt.Println(version.Version)
fmt.Println(version.GitCommit)
fmt.Println(version.BuildDate)
```

---

### CLI Usage Example

You can easily expose it via a CLI command (example using cobra):

```go
Run: func(cmd *cobra.Command, _ []string) {
  fmt.Println(version.String())
}
```

Optional short flag:

```go
if short {
  fmt.Println(version.Short())
  return
}
```

See a full usage example in the project [wis2-ingest](https://github.com/ghdrope/wis2-ingest/blob/main/cmd/version.go)

---

## Build Info Injection

These values are typically injected at build using `ldflags`:

```bash
go build -ldflags "
  -X github.com/ghdrope/go-version.Version=v1.0.0
  -X github.com/ghdrope/go-version.GitCommit=abc123
  -X github.com/ghdrope/go-version.BuildDate=2026-01-01"
```
