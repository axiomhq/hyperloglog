# hlltc
hlltc is an implementation of HyperLogLog-TailCut cardinality estimation algorithm in go.

It uses 4 bits per register instead of 5 (HyperLogLog) and 6 (HyperLogLog++), **theoretically saves 20% - 33% space**.
This implementation **practically saves 20% - 50% space** since most implementations use 1 byte registers out of convinience.


## TODO:
* [ ] more unit test coverage
* [ ] merging ability 
* [ ] marshalling and unmarshalling
