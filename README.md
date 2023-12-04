![Greenbone Logo](https://www.greenbone.net/wp-content/uploads/gb_new-logo_horizontal_rgb_small.png)

# OpenSight GoLang libraries

[![GitHub releases](https://img.shields.io/github/release/greenbone/opensight-golang-libraries.svg)](https://github.com/greenbone/opensight-golang-libraries/releases)

## About

The code maintained in this repository is used by the Greenbone OpenSight Backend Components using GoLang.

This includes
* `configReader` - reads the configuration based on environment variables with predefined defaults
* `query` - provides basic selector and response objects, including filter, paging and sorting
* `jobQueue` - a simple job queue
* `openSearch` - extensions funcs to query openSearch with the help of the query package
* `postgres` - provides a encryption / decryption package for saving data

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
