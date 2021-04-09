#!/bin/bash

#./docker_build.bash $(cat VERSION) $(cat client/VERSION)

tag="${1:?no version tag given}"
client_tag="${2:?no version tag given for client}"

tagAndPush() {
  docker tag "$1" "$2"
  docker push "$2"
}

isThere() {
  if docker pull "$1" >&- 2>&-; then
    echo 1
  else
    echo 0
  fi
}

ldIsThere="$(isThere mjuul/ld:"$tag")"

if (( ! ldIsThere )); then
  docker build -t base -f dockerfiles/Dockerfile_base .
  docker build -t ld -f dockerfiles/Dockerfile_ld .

  if tagAndPush ld mjuul/ld:"$tag"; then
    echo "pushed ld:$tag to docker-hub"
  fi
fi

ldClientIsThere="$(isThere mjuul/ld-client:"$client_tag")"

if (( ! ldClientIsThere )); then
  docker build --build-arg VERSION="$client_tag" -t ld-client -f dockerfiles/Dockerfile_client .

  if tagAndPush ld-client mjuul/ld-client:"$client_tag"; then
    echo "pushed ld-client:$client_tag to docker-hub"
  fi

  if tagAndPush ld-client mjuul/ld-client:latest; then
    echo "pushed ld-client:latest to docker-hub"
  fi

  docker build --build-arg LD_VERSION="$tag" -t ldwclient -f dockerfiles/Dockerfile_ldwclient .

  if tagAndPush ldwclient mjuul/ld:"$tag"-client; then
    echo "pushed ld: to docker-hub"
  fi
fi

if (( ldIsThere )) && (( ldClientIsThere )); then
  echo "nothing built"
  exit 1
fi

