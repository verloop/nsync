FROM alpine:3.5

ADD deploy/files/nSync /root

ENTRYPOINT ["/root/nSync"]