#!/bin/bash

tag="${1:?no version tag given}"
client_tag="${2:?no version tag given for client}"

docker build -t base -f dockerfiles/Dockerfile_base .
docker build -t ld -f dockerfiles/Dockerfile_ld .
docker build -t ld-client -f dockerfiles/Dockerfile_client .
docker build -t ldwclient -f dockerfiles/Dockerfile_ldwclient .

docker tag ld mjuul/ld:"$tag"
docker tag ld-client mjuul/ld-client:"$client_tag"
docker tag ldwclient mjuul/ld:"$tag"wclient

docker push mjuul/ld:"$tag"
docker push mjuul/ld-client:"$client_tag"
docker push mjuul/ld:"$tag"wclient