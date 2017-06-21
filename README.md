# hlltc
## TL;DR
An improved version of [HyperLogLog](https://en.wikipedia.org/wiki/HyperLogLog) for the count-distinct problem, approximating the number of distinct elements in a multiset.

## Long Version
hlltc is an implementation of HyperLogLog-TailCut cardinality estimation algorithm in go.

It uses 4 bits per register instead of 5 (HyperLogLog) and 6 (HyperLogLog++), **theoretically saves 20% - 33% space**.
This implementation **practically saves 20% - 50% space** since most implementations use 1 byte registers out of convinience.

This implementation uses the HLL++ sparse representation for lower cadinalities and the loglog-beta bias correction medium and high cardinalities.

A direct comparsion with the HyperLogLog++ implementation by Clark Duvall [https://github.com/clarkduvall/hyperloglog](https://github.com/clarkduvall/hyperloglog), gave the following results.

```
Exact 10, got:
	hlltc 10 (0.0000% off)
	hllpp 10 (0.0000% off)
Exact 50, got:
	hlltc 50 (0.0000% off)
	hllpp 50 (0.0000% off)
Exact 250, got:
	hlltc 250 (0.0000% off)
	hllpp 250 (0.0000% off)
Exact 1250, got:
	hlltc 1249 (0.0800% off)
	hllpp 1249 (0.0800% off)
Exact 6250, got:
	hlltc 6249 (0.0160% off)
	hllpp 6250 (0.0000% off)
Exact 31250, got:
	hlltc 31338 (0.2816% off)
	hllpp 31372 (0.3904% off)
Exact 156250, got:
	hlltc 157302 (0.6733% off)
	hllpp 157285 (0.6624% off)
Exact 781250, got:
	hlltc 774560 (0.8563% off)
	hllpp 774560 (0.8563% off)
Exact 3906250, got:
	hlltc 3905562 (0.0176% off)
	hllpp 3905577 (0.0172% off)
Exact 10000000, got:
	hlltc 10055418 (0.5542% off)
	hllpp 10055522 (0.5552% off)
```

## TODO:
* [ ] more unit test coverage
* [ ] merging ability 
* [ ] marshalling and unmarshalling
* [ ] benchmarks
* [ ] documentation
