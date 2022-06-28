build:
	go build -o ./cmd/gophermart/gophermart ./cmd/gophermart && chmod +x ./cmd/gophermart/gophermart

tests: build
	./cmd/tests/gophermarttest \
		-test.v -test.run=^TestGophermart$ \
		-gophermart-binary-path=cmd/gophermart/gophermart \
		-gophermart-host=localhost \
		-gophermart-port=8080 \
		-gophermart-database-uri="postgres://postgres:postgres@localhost:5455/gophermart?sslmode=disable" \
		-accrual-binary-path=cmd/accrual/accrual_linux_amd64 \
		-accrual-host=localhost \
		-accrual-port=9999 \
		-accrual-database-uri="postgres://postgres:postgres@localhost:5455/accrual?sslmode=disable"