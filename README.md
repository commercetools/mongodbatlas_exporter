# MongoDB Atlas exporter for Prometheus

## Functionality

This exporter uses [MongoDB Atlas Monitoring API](https://docs.atlas.mongodb.com/reference/api/monitoring-and-logs/) to discover Atlas resources and collect metrics from them. Metrics are collected as follow:
 * discover clusters in Atlas Project (or use the predefined by `--atlas.cluster`)
 * discover all processes (cluster nodes) of the cluster
 * collect all metrics exposed by [Get Measurements for a MongoDB Process API](https://docs.atlas.mongodb.com/reference/api/process-measurements/#measurement-values), convert and expose them
 * discover all disks belonging to the processes
 * collect all metrics exposed by [Get Measurements of a Disk for a MongoDB Process API](https://docs.atlas.mongodb.com/reference/api/process-disks-measurements/#measurement-values), convert and expose them

## Limitations

By default Atlas API allows [up to 100 requests per minute per project](https://docs.atlas.mongodb.com/api/#rate-limiting), which can be increased on personal request to your account manager/MongoDB Atlas support.

With the default API rate limit the following limitations are implied:
- Exporter supports up to 30 processes (including both mongod and mongos process types)

> Example of process number calculation:\
> 1 sharded cluster with 3 shards = 3x3 shards mongod processes + 3x3 mongos processes + 3x1 config mongod processes = 21\
> 1 non-sharded cluster (replica set) = 3x1 mongod processes

- Minimal scrape interval should be 1m

## Configuration
mongodbatlas_exporter doesn't require any configuration file and the available flags can be found as below:
```
usage: mongodbatlas_exporter [<flags>]

Flags:
  --help                    Show context-sensitive help (also try --help-long and --help-man).
  --listen-address=":9905"  The address to listen on for HTTP requests.
  --atlas.public-key=ATLAS.PUBLIC-KEY
                            Atlas API public key
  --atlas.private-key=ATLAS.PRIVATE-KEY
                            Atlas API private key
  --atlas.project-id=ATLAS.PROJECT-ID
                            Atlas project id (group id) to scrape metrics from
  --atlas.cluster=ATLAS.CLUSTER ...
                            Atlas cluster name to scrape metrics from. Can be defined multiple times. If not defined all clusters in the project will be scraped
  --log-level=debug         Printed logs level.
  --version                 Show application version.
  ```
