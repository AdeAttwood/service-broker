FROM alpine

RUN apk add ca-certificates && update-ca-certificates
ADD service-broker /opt/service-broker/service-broker
CMD /opt/service-broker/service-broker --help
