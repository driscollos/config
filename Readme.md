![License](https://img.shields.io/badge/License-This%20repo%20is%20licensed%20under%20the%20MIT%20license%20-blue)

![Library](https://img.shields.io/badge/Library%20Package%20-This%20package%20contains%20a%20library-green)

This repo is licensed under the MIT license. Please read the full license [here](https://github.com/driscollos/config/blob/main/LICENSE.md). 

# Config

This library allows you to read configuration data from a variety of
sources. Information is automatically merged according to the priority of the
source. Sources are (in priority order):

* Commandline arguments
* Environment variables
* Yaml or Json configuration files

You can access configuration data by populating a struct or by direct access
via function calls.

## Configuration Files

You can specify a file to read from, but the following files will be examined
by default. If you have more than one of the files below, they will be merged
and can override each other according to priority. The default files are
(in priority order):

* * `env.local.json`
* * `env.local.yml`
* * `config.local.json`
* * `config.local.yml`
* * `env.json`
* * `env.yml`
* * `config.json`
* * `config.yml`
* * `config/config.json`
* * `config/config.yml`
* * `build/config.json`
* * `build/config.yml`

## Populating A Struct

You can read configuration data by populating a struct. You can make use of the following tags in your structs:

* default - set a default value if no source data is found
* required (`true`) - returns an error if no data is found for this variable
* src - override the name of the data source - if you add `src="myVar"` to any variable, it will populate from the 
environment variable or yaml or json variable `myVar`

Commandline arguments, environment variables and variables in 
yaml or json files can be either a direct case match, or all in capitals eg.
if your struct contains the variable `Name`, the environment variables `Name` and
`NAME` will be a match. If you require an exact match, use the struct
tag `literal` (set to true) to enforce an exact match.

## Code Examples

### Populating A Struct

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/driscollos/config"
    "time"
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
    c.Populate(&t)

    bytes, _ := json.Marshal(t)
    fmt.Println(string(bytes))
}
```

And the associated yaml file (in this case `env.yml`)

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
      - Name: Tom
        Attendance: 30.2
        Enrolled: n
      - Name: Henry
        Attendance: 45.82
        Enrolled: false
      - Name: Laura
        Attendance: 88.1
  History:
    ClassLength: 3 hours
    Pupils:
      - Name: Pete
        Attendance: 81.4
        Enrolled: 1
    Location: Room C4
LuckyNumbers:
  - 10
  - 21
  - 56
```

The output of this code will be: 

```json
{"Name":"John","Age":41,"Classes":{"Computer Science":{"Pupils":[{"Name":"Bob","Attendance":78.4,"Enrolled":true},{"Name":"Theresa","Attendance":81.6,"Enrolled":true},{"Name":"Jim","Attendance":80.5,"Enrolled":true},{"Name":"Tom","Attendance":30.2,"Enrolled":false},{"Name":"Henry","Attendance":45.82,"Enrolled":false},{"Name":"Laura","Attendance":88.1,"Enrolled":true}],"ClassLength":7200000000000,"Location":"Spare Classroom"},"History":{"Pupils":[{"Name":"Pete","Attendance":81.4,"Enrolled":true}],"ClassLength":10800000000000,"Location":"Room C4"}},"LuckyNumbers":[10,21,56],"LotteryPicks":[10,31,55]}
```

You can override any of the data in the yaml file by setting an environment variable (as these have higher priorty than yaml files). 
For example running this:

```shell
export Classes_Computer_Science_Pupils_0_Name="Steve"
```

will change the output of the code to this:

```json
{"Name":"John","Age":41,"Classes":{"Computer Science":{"Pupils":[{"Name":"Steve","Attendance":78.4,"Enrolled":true},{"Name":"Theresa","Attendance":81.6,"Enrolled":true},{"Name":"Jim","Attendance":80.5,"Enrolled":true},{"Name":"Tom","Attendance":30.2,"Enrolled":false},{"Name":"Henry","Attendance":45.82,"Enrolled":false},{"Name":"Laura","Attendance":88.1,"Enrolled":true}],"ClassLength":7200000000000,"Location":"Spare Classroom"},"History":{"Pupils":[{"Name":"Pete","Attendance":81.4,"Enrolled":true}],"ClassLength":10800000000000,"Location":"Room C4"}},"LuckyNumbers":[10,21,56],"LotteryPicks":[10,31,55]}
```

Note that:

* You can populate elements of a slice by adding the integer index to your env variable
* Spaces in the name of variables eg. the class `Computer Science` should be converted to underscores eg `Computer_Science`

## Accessing Variables Directly

You can access parameters with the following type functions. Give the name of the variable you want to access; separate levels of nested fields
with an underscore eg. `Classes_Computer_Science_Pupils_0_Name`.

These functions will take environment variables and provide them in various formats.

* `Bool(param string) bool`
* `Date(param string) time.Time`
* `Exists(name string) bool`
* `Float(param string) float64`
* `Int(param string) int`
* `IntWithDefault(param string, defaultVal int) int`
* `String(param string) string`
* `StringWithDefault(param, defaultVal string) string`

## Duration Supported Formats

Parsing of `time.Duration` default values in struct tags supports a variety of conventions. All of the following are supported defaults:

* `1s1m1h1d`
* `1s, 1m, 1h, 1d`
* `1 second, 1 minute, 1 hour, 1 day`
* `1 sec, 1 minute, 1 hr, 1d`
* `1 second, 1 min, 1hr, 1 day`

## Specifying a file to source data from

You can specify the exact file which should be used to populate your config. If you specify a source file, all other sources are
ignored including commandline arguments and environment variables. Here is an example:

```go
c := config.New()
c.Source("./env.yml")
```