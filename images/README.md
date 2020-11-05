# Kafka Docker Image


This directory hosts the Dockerfile for the mesosphere/kafka images.

To build the image locally you can run:

```
./build.sh
```


For CI and pushing the image to DockerHub run:

```
./build.sh push
```

You likely won't have permissions to push to Dockerhub though. Use the [TeamCity job](https://teamcity.mesosphere.io/buildConfiguration/Frameworks_DataServices_Kudo_Kafka_Tools_DockerPush) to build and push instead.
