#etcd-upstream-healthchecker
```
This is a program that is used to do the healthcheck of service.
Based on service etcd registry.

Cooperation with lua-resty-upstream-etcd. 
```

##Usage
```
Usage of ./etcd-upstream-healthcheck:
  -c string
    The config file path. (default "./default.yml")
```

##Config
```
# The etcd cluster endpoints
etcdendpoints:
  - "http://172.16.1.1:2379"
  - "http://172.16.1.2:4001"

# The service regist dir in etcd
servicedir: "/v1/test/services"

# Default check url
defaultcheckurl: "/"

# When to check, exceed the timeout as fail, by ms
checktimeout: 1000

# Check interval by ms
checkinterval: 5000

# How many checks run at a time
concurrency: 50

# How many fails to set the peer down
maxfails: 3

# Not in use.
retrydelay: 1000

# Which http status will be OK?
okstatus:
  - 200
  - 201
  - 301
  - 302
  - 404
  - 400
```
