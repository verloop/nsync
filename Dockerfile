FROM golang:1.10 as build
WORKDIR /go/src/github.com/verloop/nsync
ADD . /go/src/github.com/verloop/nsync
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure
RUN CGO_ENABLED=0 go build -ldflags "-s" -a -installsuffix cgo -o /go/bin/nsync .

FROM gcr.io/distroless/base
COPY --from=build /go/bin/nsync /
CMD ["/nsync"]
