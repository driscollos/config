![License](https://img.shields.io/badge/License-MIT-blue)
![Library](https://img.shields.io/badge/Package-Library-green)
[![Go Reference](https://pkg.go.dev/badge/github.com/driscollos/config.svg)](https://pkg.go.dev/github.com/driscollos/config)

This repo is licensed under the MIT license. Please read the full license [here](https://github.com/driscollos/config/blob/main/LICENSE.md).

# Config

`config` is a Go library for loading configuration from multiple sources, with predictable precedence and hot-reload support.

Sources (highest priority first):

1. **Command-line arguments**
2. **Environment variables**
3. **YAML or JSON config files**

You can either **populate a struct** using tags or **query values directly** by path.

---

## Install

```bash
go get github.com/driscollos/config@latest
```

Import:

```go
import "github.com/driscollos/config"
```

---

## Why use `config`?

- **Clear precedence**: CLI > ENV > files (merged in priority order).
- **Hot reload**: automatically rehydrate structs or run callbacks when files change.
- **Flexible struct tags**: `default`, `required`, `src`, `base64`, `literal`.
- **Slice indexing and nested maps** supported in both YAML/JSON and env vars.
- **Duration parsing** supports natural formats like `1h`, `2 hours`, `30m`.
- **No spf13/viper dependency** — lightweight, focused, MIT-licensed.

---

## Configuration files

If you don’t specify a file, `config` searches for these defaults (in priority order):

- `env.local.json`
- `env.local.yml`
- `config.local.json`
- `config.local.yml`
- `env.json`
- `env.yml`
- `config.json`
- `config.yml`
- `config/config.json`
- `config/config.yml`
- `build/config.json`
- `build/config.yml`

Multiple files are merged; later files override earlier ones.

---

## Populating a struct

You can annotate struct fields with tags:

- `default:"value"` → fallback if no source found.
- `required:"true"` → error if no data found.
- `src:"NAME"` → override source key name.
- `base64:"optional|true"` → decode base64:
    - `optional`: decode if valid, otherwise use raw string.
    - `true`: must decode, else empty string.
- `literal:"true"` → enforce exact case match (by default keys match case-insensitive).

### Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "time"

    "github.com/driscollos/config"
)

type Teacher struct {
    Name    string `required:"true"`
    Age     int
    Classes map[string]struct {
        Pupils []struct {
            Name       string
            Attendance float64
            Enrolled   bool `default:"true"`
        }
        ClassLength time.Duration
        Location    string `default:"Spare Classroom"`
    }
    LuckyNumbers []float64
    LotteryPicks []float64 `default:"10,31,55"`
}

func main() {
    t := Teacher{}
    c := config.New()
    if err := c.Populate(&t); err != nil {
        panic(err)
    }

    b, _ := json.MarshalIndent(t, "", "  ")
    fmt.Println(string(b))
}
```

**env.yml**

```yaml
Name: John
Age: 41
Classes:
  Computer Science:
    ClassLength: 2 hours
    Pupils:
      - Name: Bob
        Attendance: 78.4
        Enrolled: yes
      - Name: Theresa
        Attendance: 81.6
        Enrolled: y
      - Name: Jim
        Attendance: 80.5
        Enrolled: true
LuckyNumbers:
  - 10
  - 21
  - 56
```

---

## Environment variable overrides

Environment variables override file values. Keys are normalized:

- Spaces, dots, and hyphens → underscores.
- Upper-cased.
- Nested paths joined with `_`.

Example:

```bash
export CLASSES_COMPUTER_SCIENCE_PUPILS_0_NAME="Steve"
```

Overrides the first pupil’s name, even if originally set in YAML.

---

## Accessing variables directly

If you don’t want to bind to a struct, you can fetch values by path (underscore- or dot-separated). Examples:

```go
c := config.New()

name := c.String("Classes.Computer_Science.Pupils.0.Name")
age := c.Int("Age")
ok   := c.Bool("Classes.History.Pupils.0.Enrolled")
```

Available helpers:

- `Bool(param string) bool`
- `Date(param string) time.Time`
- `Exists(name string) bool`
- `Float(param string) float64`
- `Int(param string) int`
- `IntWithDefault(param string, defaultVal int) int`
- `String(param string) string`
- `StringWithDefault(param, defaultVal string) string`

---

## Duration formats

`time.Duration` fields accept a wide range of inputs:

- `1s1m1h1d`
- `1s, 1m, 1h, 1d`
- `1 second, 1 minute, 1 hour, 1 day`
- `1 sec, 1 min, 1 hr, 1d`

---

## Specifying a file explicitly

```go
c := config.New()
c.Source("./env.yml")
```

This disables default file discovery and loads only from the given file (plus env/CLI overrides).

---

## Hot reload

`config` can watch source files and update values at runtime.

- Files must exist at startup to be watched.
- On change, all watched files are reloaded atomically.
- You can re-hydrate structs or run callback functions.

### Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/driscollos/config"
)

type myData struct {
    Name string
}

func main() {
    c := config.New()
    data := myData{}
    if err := c.Populate(&data); err != nil {
        panic(err)
    }

    c.HotReload(context.Background(), &data, func() {
        fmt.Println("new name is:", data.Name)
    })

    for {
        fmt.Println("name", data.Name)
        time.Sleep(time.Second)
    }
}
```

---

## Contributing

- Run tests: `go test -race ./...`
- Generate mocks: `go generate ./...`
- Please open PRs or issues for bug reports and enhancements.

---

## License

MIT © [John Driscoll](https://github.com/codebyjdd). See [LICENSE.md](LICENSE.md).
