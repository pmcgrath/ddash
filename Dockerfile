FROM progrium/busybox 
MAINTAINER pmcgrath@gmail.com

COPY ddash /usr/bin/ddash

EXPOSE 8090

ENTRYPOINT ["/usr/bin/ddash"]
