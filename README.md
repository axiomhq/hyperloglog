# hlltc
hlltc is an implementation of HyperLogLog-TailCut cardinality estimation algorithm in go.

It uses 4 bits per register instead of 5 (HyperLogLog) and 6 (HyperLogLog++), **theoretically saves 20% - 33% space**.
This implementation **practically saves 20% - 50% space** since most implementations use 1 byte registers out of convinience.

This implementation uses the HLL++ sparse representation for lower cadinalities and the loglog-beta bias correction medium and high cardinalities.

A direct comparsion with the HyperLogLog++ implementation by Clark Duvall [https://github.com/clarkduvall/hyperloglog](https://github.com/clarkduvall/hyperloglog), gave the following results.

```
Exact 10, got:
	hlltc 10 (0.00% off)
	hllpp 10 (0.00% off)
Exact 50, got:
	hlltc 50 (0.00% off)
	hllpp 50 (0.00% off)
Exact 250, got:
	hlltc 250 (0.00% off)
	hllpp 250 (0.00% off)
Exact 1250, got:
	hlltc 1250 (0.00% off)
	hllpp 1250 (0.00% off)
Exact 6250, got:
	hlltc 6250 (0.00% off)
	hllpp 6251 (0.02% off)
Exact 31250, got:
	hlltc 31412 (0.52% off)
	hllpp 31288 (0.12% off)
Exact 156250, got:
	hlltc 156388 (0.09% off)
	hllpp 154084 (1.39% off)
Exact 781250, got:
	hlltc 780506 (0.10% off)
	hllpp 780176 (0.14% off)
Exact 3906250, got:
	hlltc 3938432 (0.82% off)
	hllpp 3935760 (0.76% off)
Exact 10000001, got:
	hlltc 10092457 (0.92% off)
	hllpp 10099191 (0.99% off)
```

## TODO:
* [ ] more unit test coverage
* [ ] merging ability 
* [ ] marshalling and unmarshalling
