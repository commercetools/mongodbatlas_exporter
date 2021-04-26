# mongodbatlas_exporter


## Limitations

- Exporter supports up to 30 processes (mongod and mongos)
> number of process calculation:\
> 1 sharded cluster with 3 shards = 3x3 shards mongod processes + 3x3 mongos processes + 3x1 config mongod processes = 21\
> 1 non-sharded cluster (replica set) = 3x1 mongod processes
- Minimal scrape interval should be 1m
