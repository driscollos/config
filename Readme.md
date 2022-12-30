![License](https://img.shields.io/badge/License-This%20repo%20is%20licensed%20under%20the%20MIT%20license%20-blue)

![Library](https://img.shields.io/badge/Library%20Package%20-This%20package%20contains%20a%20library-green)

This repo is licensed under the MIT license. Please read the full license [here](https://github.com/driscollos/config/blob/main/LICENSE.md). 

# Parameters

This library allows you to read parameters from a variety of sources; environment variables, commandline arguments and yaml configuration
files. You can either access variables by function call, or populate a configuration struct.

Configuration structs can be nested to as many levels as you like.

Values are sourced from various places. These are the sources, in order of precedence with the highest priority first.

* Commandline arguments eg `--name John`
* Environment variables
* Yaml files
* * `env.local.yml`
* * `config.local.yml`
* * `env.yml`
* * `config.yml`
* * `config/config.yml`
* * `build/config.yml`
* default values in struct tags eg `default:"MyDefaultValue"`

You can specify a particular yaml file to use as the exclusive source of values using the `Source(filename)`
function, which is a method of the `config` struct.

You can also set a struct tag `required` which will result in an error if no value can be found for that variable and no default value is defined.

## Code Examples

### Populating A Struct

```go
import (
	"encoding/json"
	"fmt"
	"github.com/driscollos/config"
	"time"
)

func main() {
    conf := config.New()
    mine := myConf{}
    if err := conf.Populate(&mine); err != nil {
        fmt.Println("could not populate configuration:", err.Error())
        os.Exit(0)
    }

    bytes, _ := json.Marshal(mine)
    fmt.Println(string(bytes))
}

type myConf struct {
    Name        string
    Age         int       `default:"40"`
    TimeOfBirth time.Time `default:"1981-12-01 17:21:33" layout:"2006-01-02 15:04:05"`
    Hobbies     struct {
        First  string `default:"sports"`
        Second string `default:"coding"`
        Best   struct {
            ReasonOne string
            ReasonTwo string `default:"just because"`
        }
    }
    FaveColour Colour
}

type Colour struct {
    Name string
    ID   string `default:"blue"`
}
```

The output of this code will be: 

```go
{"Name":"","Age":40,"TimeOfBirth":"1981-12-01T17:21:33Z","Hobbies":{"First":"sports","Second":"coding","Best":{"ReasonOne":"","ReasonTwo":"just because"}},"FaveColour":{"Name":"","ID":"blue"}}
```

As you can see, the `default` tags in our struct have been used to populate variables as no environment variables are set. \
Let's try creating an environment yaml file to set some values. Create the file `env.yml` and populate it like this:

```yaml
Name: John
Age: 30
Hobbies:
  First: Travel
  Best:
    ReasonOne: Because it is the best
    ReasonTwo: Do I need a second reason?
FaveColour:
  Name: Red
```

Running your code again, you will see this result:

```go
{"Name":"John","Age":30,"TimeOfBirth":"1981-12-01T17:21:33Z","Hobbies":{"First":"Travel","Second":"coding","Best":{"ReasonOne":"Because it is the best","ReasonTwo":"Do I need a second reason?"}},"FaveColour":{"Name":"Red","ID":"blue"}}
```

This shows we have populated even nested fields from our Yaml file, and overriden defaults defined in our struct.

#### Environment variables for nested fields

To set an environment variable for a nested field, use an underscore for each level eg. `export Hobbies_First="Ice Skating"` - this will override the 
default value of `Hobbies.First` defined in your struct, and also any variable set in a local environment file eg. `env.yml`.

## Accessing Variables Directly

You can access parameters with the following type functions. Give the name of the variable you want to access; separate levels of nested fields
with an underscore eg. `Hobbies_First`.

These functions will take environment variables and provide them in various formats.

* `Bool(param string) bool`
* `Date(param string) time.Time`
* `Exists(name string) bool`
* `Float(param string) float64`
* `Int(param string) int`
* `IntWithDefault(param string, defaultVal int) int`
* `String(param string) string`
* `StringWithDefault(param, defaultVal string) string`

## Specifying a file to source data from

You are able to specify a source yaml file that config should use as the exclusive source of information.
Take a look at this example:

```go
package main

import (
    "fmt"
    "github.com/driscollos/config"
    "os"
)

func main() {
    appConf := myApplicationConfig{}
    c := config.New()
    c.Source("example.yaml")
    err := c.Populate(&appConf)
    if err != nil {
        fmt.Println("error:", err.Error())
        os.Exit(0)
    }
    fmt.Println(appConf.DownstreamService.Enabled, appConf.DownstreamService.Address)
}

type myApplicationConfig struct {
    DownstreamService struct {
        Enabled bool   `default:"true"`
        Address string `required:"true"`
    }
}
```

And the file `example.yaml`
```yaml
DownstreamService:
  Address: https://example.org
```

The output from this will be `true https://example.org`