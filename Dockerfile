FROM alpine:3

# We then install git and the package required to build llama and the go app
RUN apk add go make git g++ ffmpeg

# And we finally build the app
WORKDIR /go/src/github.com/polyfact/api

COPY . /go/src/github.com/polyfact/api/

RUN go get

RUN make build/server_start

ENTRYPOINT /go/src/github.com/polyfact/api/build/server_start
