FROM ubuntu:latest
ENV ADDR=localhost:8088
ENV DATABASE_URI=postgres://postgres:postgres@localhost:5455/accrual?sslmode=disable
COPY ./cmd/accrual/accrual_linux_amd64 /accural
CMD /accural -a $ADDR -d $DATABASE_URI