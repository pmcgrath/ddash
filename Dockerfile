FROM progrium/busybox 
MAINTAINER pmcgrath@gmail.com

ADD ddash /usr/bin/ddash

EXPOSE 8090

ENTRYPOINT ["/usr/bin/ddash"]
