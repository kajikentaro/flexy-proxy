cpu-profile:
	go test ./integration_test/benchmark_test -bench=. -cpuprofile=cpu.out
	go tool pprof -http=0.0.0.0:8080 benchmark_test.test cpu.out

mem-profile:
	go test ./integration_test/benchmark_test -bench=. -benchmem -memprofile=mem.out -cpuprofile=cpu.out
	go tool pprof -http=0.0.0.0:8080 benchmark_test.test mem.out
