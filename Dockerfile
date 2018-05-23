FROM golang:1.10 as build
WORKDIR /go/src/app
ADD . /go/src/app
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure
RUN go install .

FROM gcr.io/distroless/base
COPY --from=build /go/bin/app /
ENTRYPOINT ["/app"]
