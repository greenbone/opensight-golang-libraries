# Config Reader

The `configReader` package provides a function `ReadEnvVarsIntoStruct` that can be used to read environment variables or config files into a struct.

## Usage

To use the `configReader` package, first create a struct that you want to read environment variables into:

```go
type Config struct {
	Port     int    `viperEnv:"port"`
	Host     string `viperEnV:"host" default:"localhost"`
	LogLevel string `viperEnv:"LOG_LEVEL" default:"info"`
}
```

Then, in your code, you can use the `ReadEnvVarsIntoStruct` function to read environment variables into the struct:

```go
package main

import (
	"github.com/greenbone/opensight-golang-libraries/configReader"
)

func main() {
	var config Config
	err := configReader.ReadEnvVarsIntoStruct(&config)
	if err != nil {
		panic(err)
	}
	// use the config
}
```

Any environment variables that match the struct fields will be read into the struct. The following environment variables will be read into the `Config` struct:

```
PORT=8080
HOST=localhost
LOG_LEVEL=debug
```

If a field in the struct has a `viperEnv` tag, the environment variable will be matched to that tag. Otherwise, the field name will be used as the tag.

If a field in the struct has a `default` tag, a default value will be used if the environment variable is not set.

The `configReader` package uses [Viper](https://github.com/spf13/viper) for reading environment variables, and [Zerolog](https://github.com/rs/zerolog) for logging.

## Maintainer

This project is maintained by [Greenbone AG][Greenbone AG]

## Contributing

Your contributions are highly appreciated. Please
[create a pull request](https://github.com/greenbone/asset-management-backend/pulls)
on GitHub. Bigger changes need to be discussed with the development team via the
[issues section at GitHub](https://github.com/greenbone/asset-management-backend/issues)
first.

## License

Copyright (C) 2022-2023 [Greenbone AG][Greenbone AG]

Licensed under the [GNU General Public License v3.0 or later](LICENSE).

[Greenbone AG]: https://www.greenbone.net/
[poetry]: https://python-poetry.org/
[pip]: https://pip.pypa.io/
[autohooks]: https://github.com/greenbone/autohooks