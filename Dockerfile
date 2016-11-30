FROM scratch
MAINTAINER Stepan Stipl

ADD k8s-ns-meddler /

EXPOSE 8080

CMD ["/k8s-ns-meddler"]
