# garena: simple and experimental arena allocator for Go
---

## Some Benchmarks 
![bench](img/bench.png "Benchmark on MBP M1")
![bench](img/bench_x86_64.png "Benchmark on old core i5")

* Bench 1 (single allocation speed) -> at least 77% faster than GC alloc
* Bench 2 (multiple allocation speed) -> at least 770% faster than GC alloc
