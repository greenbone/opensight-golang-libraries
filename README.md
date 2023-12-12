![Greenbone Logo](https://www.greenbone.net/wp-content/uploads/gb_new-logo_horizontal_rgb_small.png)

# OpenSight GoLang libraries

[![GitHub releases](https://img.shields.io/github/release/greenbone/opensight-golang-libraries.svg)](https://github.com/greenbone/opensight-golang-libraries/releases)

## About

The code maintained in this repository is used by the Greenbone OpenSight Backend Components using GoLang.

The following funtionalities are provided:
* [configReader](pkg/configReader/README.md) - reads the configuration based on environment variables with predefined defaults
* [dbcrypt](pkg/dbcrypt/README.md) - provides a encryption / decryption package for saving data
* [jobQueue](pkg/jobQueue/README.md) - a simple job queue
* [openSearch](pkg/openSearch) - extensions funcs to query openSearch with the help of the query package
* [query](pkg/query/README.md) - provides basic selector and response objects, including filter, paging and sorting
* [slices](pkg/query/README.md) - utility functions for slices
* [testFolder](pkg/testFolder) - access to test data from the file system

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
