FROM mesosphere/kafka:2.5.1-1.3.3
USER root
ARG UNAME=kafka

RUN apt update -qq &&\
    apt install gettext-base procps -y
RUN rm -rf /var/lib/apt/lists/*

ENV PATH $PATH:/opt/kafka/bin
COPY scripts/* /opt/kafka/

RUN chown -R "$UNAME:$UNAME" /opt/kafka/
USER $UNAME
