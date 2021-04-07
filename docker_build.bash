#!/bin/bash

tag="${1:?no version tag given}"
client_tag="${2:?no version tag given for client}"

docker build -t base -f dockerfiles/Dockerfile_base .
docker build -t ld -f dockerfiles/Dockerfile_ld .
docker build --build-arg VERSION="$client_tag" -t ld-client -f dockerfiles/Dockerfile_client .

docker tag ld mjuul/ld:"$tag"
docker tag ld-client mjuul/ld-client:"$client_tag"
docker tag ld-client mjuul/ld-client:latest

docker push mjuul/ld:"$tag"
docker push mjuul/ld-client:"$client_tag"
docker push mjuul/ld-client:latest
