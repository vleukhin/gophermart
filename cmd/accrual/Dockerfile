FROM ubuntu:latest
ENV ADDR=localhost:8088
COPY ./cmd/accrual/accrual_linux_amd64 /accural
CMD /accural -a $ADDR