# Model

The "model" package's core functionality centers around providing the necessary
bridge between `mongodbatlas.Measurements`, `mongodbatlas.Measurement`, and `prometheus`.


Any method calling atlas for `mongodbatlas.Measurements` should return a `model.Measurement` as its final result.
