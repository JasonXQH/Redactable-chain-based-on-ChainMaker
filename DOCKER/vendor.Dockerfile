FROM golang:1.18.10 as builder
ENV GOPROXY=https://goproxy.cn,direct
ENV GOPRIVATE=chainmaker.org
COPY . /chainmaker-go
#RUN cd /chainmaker-go && make vendor-build && make vendor-build-cmc
RUN cd /chainmaker-go && make vendor-build

# the second stage
FROM ubuntu:20.04
RUN rm /bin/sh && ln -s /bin/bash /bin/sh
RUN sed -i "s@http://.*archive.ubuntu.com@http://mirrors.tuna.tsinghua.edu.cn@g" /etc/apt/sources.list && \
    sed -i "s@http://.*security.ubuntu.com@http://mirrors.tuna.tsinghua.edu.cn@g" /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y vim net-tools tree gcc g++ p7zip-full

ENV TZ "Asia/Shanghai"
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata && \
    echo $TZ > /etc/timezone && \
    ln -fs /usr/share/zoneinfo/$TZ /etc/localtime && \
    dpkg-reconfigure tzdata -f noninteractive

COPY --from=builder /chainmaker-go/main/libwasmer_runtime_c_api.so /usr/lib/libwasmer.so
COPY --from=builder /chainmaker-go/main/prebuilt/linux/wxdec /usr/bin
COPY --from=builder /chainmaker-go/bin/chainmaker /chainmaker-go/bin/
COPY --from=builder /chainmaker-go/config /chainmaker-go/config/
RUN mkdir -p /chainmaker-go/log/
RUN chmod 755 /usr/bin/wxdec

WORKDIR /chainmaker-go/bin
