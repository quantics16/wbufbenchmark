# wbufbenchmark
Benchmarks for https://github.com/jackc/pgconn/issues/137

### Direct Test
Testing the method is believed to allocate memory https://github.com/jackc/pgconn/blob/master/pgconn.go#L1110. The thought is that every time these encode methods are run, if the buffer isn't large enough it causes allocations. And if the overall size is sufficiently large enough, it could be a significant amount of allocations.

Testing using a multi-insert query containing 1000 values, each of which are roughly 16KB in size, for a total of ~16MB in size for the query as a whole
```
go test -bench .*Direct.*
goos: darwin
goarch: amd64
pkg: github.com/quantics16/wbufbenchmark
cpu: VirtualApple @ 2.50GHz
BenchmarkPgConnWbufDirect1KB-10     	      27	  37558035 ns/op	88242940 B/op	      36 allocs/op
BenchmarkPgConnWbufDirect1MB-10     	      32	  38572945 ns/op	91922612 B/op	      14 allocs/op
BenchmarkPgConnWbufDirect20MB-10    	    1186	    930651 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/quantics16/wbufbenchmark	4.598s
```

The values for this test were fetched from dumping the contents of the pgproto3.Bind object that was provided from the live test.

### Live Test
Testing the method against a live postgres instance. Intent is to find the relative performance difference between runs with a different wbuflen value, since the benchmark otherwise includes the entire req/resp from postgres.

Reqiuires a locally running postgres instance, then it uses the following env vars to connect to it
```
PQL_USER
PQL_PASSWORD
PQL_DATABASE
```

Tested using a multi-insert query containing 1000 values, each of which are roughly 16KB in size, for a total of ~16MB in size for the query as a whole
```
go test -bench .*ViaPostgres.*
message size in bytes: 16507576
goos: darwin
goarch: amd64
pkg: github.com/quantics16/wbufbenchmark
cpu: VirtualApple @ 2.50GHz
BenchmarkPgConnWBufViaPostgres-10    	       1	1035001708 ns/op	107349512 B/op	    5079 allocs/op
PASS
ok  	github.com/quantics16/wbufbenchmark	1.698s
```

I used vendoring (go mod vendor) to make it easy to change wbuflen, but otherwise modifying the value in the file manually to 1MB..
```
go test -bench .*ViaPostgres.*
message size in bytes: 16507585
goos: darwin
goarch: amd64
pkg: github.com/quantics16/wbufbenchmark
cpu: VirtualApple @ 2.50GHz
BenchmarkPgConnWBufViaPostgres-10    	       1	1081868042 ns/op	110988280 B/op	    5057 allocs/op
PASS
ok  	github.com/quantics16/wbufbenchmark	1.771s
```

Doing the same, but this time to 20MB, since then the buffer is larger than the entire query..
```
go test -bench .*ViaPostgres.*
message size in bytes: 16507626
goos: darwin
goarch: amd64
pkg: github.com/quantics16/wbufbenchmark
cpu: VirtualApple @ 2.50GHz
BenchmarkPgConnWBufViaPostgres-10            2	 745823666 ns/op	19029052 B/op	    5032 allocs/op
PASS
ok  	github.com/quantics16/wbufbenchmark	2.918s
```

or for a side by side view
```
BenchmarkPgConnWBufViaPostgres1KB-10    	       1	1035001708 ns/op	107349512 B/op	    5079 allocs/op
BenchmarkPgConnWBufViaPostgres1MB-10    	       1	1081868042 ns/op	110988280 B/op	    5057 allocs/op
BenchmarkPgConnWBufViaPostgres20MB-10    	       2	 745823666 ns/op	19029052 B/op	    5032 allocs/op
```

