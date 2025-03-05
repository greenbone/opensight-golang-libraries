![Greenbone Logo](https://www.greenbone.net/wp-content/uploads/gb_new-logo_horizontal_rgb_small.png)

# ginSwagger

provides a safe way to server the SwaggerUI

## Reason for implementation
Reason to provide this package is that the https://github.com/swaggo/gin-swagger is actually not using the latest swagger-ui.

To use the latest swagger-ui (by using github.com/swaggo/files/v2) we needed to change the implementation as the gin-swagger package sitll uses the files in version 1.0

## Sources used
To compile the new package we took the gin-echo package (https://github.com/swaggo/echo-swagger/blob/master/swagger.go) as a reference and added the gin based functionality.


## Usage

If you need the possibility to authenticate using keycloak you first need to set the OAuthConfig.
Then provide this OAuthConfig to the ginSwagger.GinWrapHandler which serves the swagger UI.

Example
```go
authConfig := &ginSwagger.OAuthConfig{
    ClientId: cfg.WebClientName,
    Realm:    cfg.Realm,
    AppName:  "Asset Management Backend",
}

ginSwagger.GinWrapHandler(
    ginSwagger.OAuth(authConfig),
    ginSwagger.InstanceName(""),
)(c)
```

# License

Copyright (C) 2022-2023 [Greenbone AG][Greenbone AG]

Licensed under the [GNU General Public License v3.0 or later](../../LICENSE).