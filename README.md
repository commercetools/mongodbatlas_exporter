# Mongodbatlas exporter for Prometheus

## Limitations

- Exporter supports up to 30 processes (mongod and mongos)
> number of process calculation:\
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
