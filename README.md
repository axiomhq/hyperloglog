# hlltc
An improved version of [HyperLogLog](https://en.wikipedia.org/wiki/HyperLogLog) for the count-distinct problem, approximating the number of distinct elements in a multiset. 

This work is based on ["Better with fewer bits: Improving the performance of cardinality estimation of large data streams - Qingjun Xiao, You Zhou, Shigang Chen"](http://cse.seu.edu.cn/PersonalPage/csqjxiao/csqjxiao_files/papers/INFOCOM17.pdf).

## Implementation
hlltc is an implementation of HyperLogLog-TailCut cardinality estimation algorithm in go.

The core difference to other implementations are:
* **use metro hash** instead of xxhash
* **sparse representation** for lower cadinalities and the loglog-beta bias correction medium and high cardinalities.
* **4-bit register** instead of 5 (HyperLogLog) and 6 (HyperLogLog++), but most implementations use use 1 byte registers out of convinience, thus **practically saves 20% - 50% space**.

This implementation uses the HLL++ sparse representation for lower cadinalities and the loglog-beta bias correction medium and high cardinalities. In general it borrows a lot from the [InfluxData's fork](https://github.com/influxdata/influxdb/tree/master/pkg/estimator/hll) of [Clark Duvall HyperLogLog++ implementation](https://github.com/clarkduvall/hyperloglog).

## Results
A direct comparsion with the [HyperLogLog++ implementation by Clark Duvall](https://github.com/clarkduvall/hyperloglog), yielded the following results.

| Exact | HLLPP | HLLTC |
| --- | --- | --- |
| 10 | 10 (0.0% off) | 10 (0.0% off) |
| 50 | 50 (0.0% off) | 50 (0.0% off) |
| 250 | 250 (0.0% off) | 250 (0.0% off) |
| 1250 | 1249 (0.08% off) | 1249 (0.08% off) |
| 6250 | **6250 (0.0% off)** | 6249 (0.016% off) |
| 31250 | 31372 (0.3904% off) | **31338 (0.2816% off)** |
| 156250 | **157285 (0.6624% off)** | 157302 (0.6733% off) |
| 781250 |  774560 (0.8563% off) | 774560 (0.8563% off) |
| 3906250 | **3905577 (0.0172% off)** | 3905562 (0.0176% off) |
| 10000000 | 10055522 (0.5552% off) | **10055418 (0.5542% off)** |


## Note
A big thank you to Prof. Shigang Chen and his team at the University of Florida who are actively conducting research around "Big Network Data".

## TODO:
* [ ] more unit test coverage
* [ ] merging ability 
* [ ] benchmarks
* [ ] documentation
