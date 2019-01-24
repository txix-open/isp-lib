# ISP-LIB

## Module configuration
Local configuration loads on module startup. To specify path to local config file set `APP_CONFIG_PATH` environment variable. By default in production mode `config.yml` loads.

To override configuration from file, set environment variables with prefix `ISP_` and upper case property name.
```yml
configServiceAddress:
  ip: 127.0.0.1
  port: 9001
```
```
ISP_CONFIGSERVICEADDRESS.IP=172.16.0.1
ISP_CONFIGSERVICEADDRESS.PORT=9000
```
