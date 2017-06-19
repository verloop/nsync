FROM alpine:3.5

ADD deploy/files/nsync /root

ENTRYPOINT ["/root/nsync"]