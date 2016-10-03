<p align="center">
<img 
    src="assets/logo.png" width="240" height="78" border="0" alt="DJSON">
<br/><br/>
<a href="https://godoc.org/github.com/a8m/djson"><img src="https://img.shields.io/badge/api-reference-blue.svg?style=flat-square" alt="GoDoc"></a>
<a href="https://travis-ci.org/a8m/djson"><img src="https://img.shields.io/travis/a8m/djson.svg?style=flat-square"
alt="Build Status"></a>
<a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="LICENSE"></a>
</p>

DJSON is a JSON decoder for Go that is ___2~ to 3~ times faster___ than
the standard `encoding/json` and the existing solutions, when dealing with
arbitrary JSON payload. [See benchmarks below](#benchmark).  
It is a good approach for people who are using `json.Unmarshal` together
with `interface{}`, don't know what the schema is, and still want good
performance with minimal changes.

### Motivation
While searching for a JSON parser solution for my projects, that is faster than the standard library, with zero reflection tests, allocates less memory and is still safe(I didn't want the `"unsafe"` package in my production code, in order to reduce memory consumption).  
I found that almost all implemtations are just wrappers around the standard library
and aren't fast enough for my needs.  
I encountered two projects: [ujson](https://github.com/mreiferson/go-ujson) that is the UltraJSON implementation
and [jsonparser](https://github.com/buger/jsonparser), that is a pretty awesome project.  
ujson seems to be faster than `encoding/json` but still doesn't meet my requirements.  
jsonparser seems to be really fast, and I even use it for some of my new projects.  
However, its API is different, and I would need to change too much of my
code in order to work with it.  
Also, for my processing work that involves `ETL`, changing and setting new
fields on the JSON object, I need to transform the `jsonparser`
result to `map[string]interface{}` and it seems that it loses its power.

### Advantages and Stability
As you can see in the [benchmark below](#benchmark), DJSON is faster and allocates less
memory than the other alternatives.  
The current version is `1.0.0-alpha.1`, and I'm waiting to hear from you
if there are any issues or bug reports, to make it stable.  
(comment: there is a test file named `decode_test` that contains a [test case](https://github.com/a8m/djson/blob/master/decode_test.go#L104) that
compares the results to `encoding/json` - feel free to add more values if you find they are important)  
I'm also plaining to add the `DecodeStream(io.ReaderCloser)` method(or `NewDecoder(io.ReaderCloser)`), to support stream decoding
without breaking performance.


### Benchmark
There are 3 benchmark types: [small](#small-payload), [medium](#medium-payload) and [large](#large-payload) payloads.  
All the 3 are taken from the `jsonparser` project, and they try to simulate a real-life usage. 
Each result from the different benchmark types is shown in a metric table below.
The lower the metrics are, the better the result is.
__Time/op__ is in nanoseconds, __B/op__ is how many bytes were allocated
per op and __allocs/op__ is the total number of memory allocations.  
Benchmark results that are better than `encoding/json` are marked in bold text.  
The Benchmark tests run on AWS EC2 instance(c4.xlarge). see: [screenshots](https://github.com/a8m/djson/tree/master/assets)

Compared libraries:
- https://golang.org/pkg/encoding/json
- https://github.com/Jeffail/gabs
- https://github.com/bitly/go-simplejson
- https://github.com/antonholmquist/jason
- https://github.com/mreiferson/go-ujson
- https://github.com/ugorji/go/codec

#### Small payload
Each library in the test gets a small payload to process that weighs 134 bytes.  
You can see the payload [here](https://github.com/a8m/djson/blob/master/benchmark/benchmark_fixture.go#L3), and the test screenshot [here](https://github.com/a8m/djson/blob/master/assets/bench_small.png).

| __Library__                 | __Time/op__   | __B/op__ | __allocs/op__ |
|-----------------------------|-------------- |----------|---------------|
| encoding/json               |    8646       |   1993   |   60          |
| ugorji/go/codec             |    9272       |   4513   |   __41__      |
| antonholmquist/jason        |    __7336__   |   3201   |   __49__      |
| bitly/go-simplejson         |    __5253__   |   2241   |   __36__      |
| Jeffail/gabs                |    __4788__   | __1409__ |   __33__      |
| mreiferson/go-ujson         |    __3897__   | __1393__ |   __35__      |
| a8m/djson                   |    __2534__   | __1137__ |   __25__      |
| a8m/djson.[AllocString][as] |    __2195__   | __1169__ |   __13__      |


#### Medium payload
Each library in the test gets a medium payload to process that weighs 1.7KB.  
You can see the payload [here](https://github.com/a8m/djson/blob/master/benchmark/benchmark_fixture.go#L5), and the test screenshot [here](https://github.com/a8m/djson/blob/master/assets/bench_medium.png).

| __Library__                  | __Time/op__    | __B/op__ | __allocs/op__  |
|------------------------------|----------------|-----------|---------------|
| encoding/json                |    42029       |   10652   |   218         |
| ugorji/go/codec              |    65007       |   15267   |   313         |
| antonholmquist/jason         |    45676       |   17476   |   224         |
| bitly/go-simplejson          |    45164       |   17156   |   219         |
| Jeffail/gabs                 |    __41045__   | __10515__ |   __211__     |
| mreiferson/go-ujson          |    __33213__   |   11506   |   267         |
| a8m/djson                    |    __22871__   | __10100__ |   __195__     |
| a8m/djson.[AllocString][as]  |    __19296__   | __10619__ |   __87__      |

#### Large payload
Each library in the test gets a large payload to process that weighs 28KB.  
You can see the payload [here](https://github.com/a8m/djson/blob/master/benchmark/benchmark_fixture.go#L7), and the test screenshot [here](https://github.com/a8m/djson/blob/master/assets/bench_large.png).

| __Library__                 | __Time/op__    | __B/op__   | __allocs/op__  |
|-----------------------------|----------------|------------|----------------|
| encoding/json               |    717882      |   212827   |   3247         |
| ugorji/go/codec             |    1052347     |   239130   |   4426         |
| antonholmquist/jason        |    751910      |   277931   |   3257         |
| bitly/go-simplejson         |    753663      |   277628   |   3252         |
| Jeffail/gabs                |    __714304__  | __212740__ |   __3241__     |
| mreiferson/go-ujson         |    __599868__  |   235789   |   4057         |
| a8m/djson                   |    __437031__  | __210997__ |   __2932__     |
| a8m/djson.[AllocString][as] |    __372382__  |   214053   |   __1413__     |


### LICENSE
MIT

[as]: https://github.com/a8m/djson/blob/master/decode.go#L25
