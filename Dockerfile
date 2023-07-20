FROM alpine:3

RUN apk add go make git

ENV GOPATH=/go

RUN mkdir -p $GOPATH/src/github.com/tmc/

WORKDIR /go/src/github.com/tmc

RUN git clone https://github.com/lowczarc/langchaingo.git

WORKDIR /go/src/github.com/tmc/langchaingo

RUN git checkout 82a0238

WORKDIR /go/src/github.com/polyfact/api

COPY . /go/src/github.com/polyfact/api/

RUN echo "replace github.com/tmc/langchaingo => $GOPATH/src/github.com/tmc/langchaingo" >> go.mod

RUN go get

RUN make build/server_start

ENTRYPOINT /go/src/github.com/polyfact/api/build/server_start
