#!/usr/bin/env bash

BASEDIR=$(dirname "$0")
PROJECT_DIR="$(realpath "${BASEDIR}")"

IMAGE_NAME="go-space-chat"

GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
GIT_COMMIT=$(git rev-parse --short HEAD)
IMAGE_TAG="${GIT_BRANCH}-${GIT_COMMIT}"

function run_docker() {
    echo "Running docker image: $IMAGE_NAME:$IMAGE_TAG"
    docker run -it --rm \
        -w /app \
        -p 8081:80 \
        -p 9000:9000 \
        $IMAGE_NAME:$IMAGE_TAG \
        /app/go-space-chat
}

# rm docker containers and images if exists
CONTAINERS=$(docker ps -a -q -f name=$IMAGE_NAME)
if [ -n "$CONTAINERS" ]; then
    docker rm -f $CONTAINERS
fi

IMAGES=$(docker images -q $IMAGE_NAME)
if [ -n "$IMAGES" ]; then
    run_docker

    exit 0
fi

echo "Will build image: $IMAGE_NAME:$IMAGE_TAG"
# build image
docker build --progress=plain --no-cache -t $IMAGE_NAME:$IMAGE_TAG -f ./Dockerfile ./

echo "Image $IMAGE_NAME:$IMAGE_TAG is Built!"

run_docker
