FROM golang:alpine AS build-env
WORKDIR /etc
RUN mkdir bootstrap
ADD ./bootstrap ./bootstrap
RUN apk add --no-cache gcc musl-dev git mercurial ca-certificates
ENV GO111MODULE=on
RUN cd ./bootstrap && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap


# using debian:stretch-20190610-slim
FROM debian@sha256:9490c476443a3869e39c2897fa66c91daf5dcbbfca53c976dac7bbdc45775b28

RUN apt-get update && apt-get install -y curl openssl netcat procps
RUN rm -rf /var/lib/apt/lists/*

ENV PIPELINE_HOME="/etc/pipeline"

ARG UNAME=kafka
ARG UID=1000
ARG GID=1000
RUN groupadd -g $GID -o $UNAME && \
    useradd -r -m -u $UID -g $GID $UNAME

WORKDIR $PIPELINE_HOME

COPY --from=build-env /etc/bootstrap/ /etc/pipeline/
RUN chown -R "$UNAME:$UNAME" $PIPELINE_HOME

USER $UNAME
