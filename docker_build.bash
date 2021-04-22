#!/bin/bash

#./docker_build.bash $(cat VERSION) $(cat client/VERSION)

tag="${1:?no version tag given}"

tagAndPush() {
  docker tag "$1" "$2"
  docker push "$2"
}

if docker pull mjuul/ld:"$tag" >&- 2>&-; then
  echo "nothing built"
  exit 1
else
  docker build -t ld -f dockerfiles/Dockerfile_ld .
  docker build -t ldwclient -f dockerfiles/Dockerfile_ldwclient .

  if tagAndPush ld mjuul/ld:"$tag"; then
    echo "pushed ld:$tag to docker-hub"
  fi

  if tagAndPush ldwclient mjuul/ld:"$tag"-client; then
    echo "pushed ld: to docker-hub"
  fi
fi


