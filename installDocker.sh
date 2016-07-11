#!/bin/sh

# Update to most recent docker version
if [[ "$TRAVIS_OS_NAME" == "linux" ]]; then
  sudo apt-get update;
  sudo apt-cache search docker;
  sudo apt-get -o Dpkg::Options::="--force-confnew" install -y docker-engine;
fi
# Docker-compose installation
sudo rm /usr/local/bin/docker-compose || true
curl -L https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-`uname -s`-`uname -m` > docker-compose
chmod +x docker-compose
sudo mv docker-compose /usr/local/bin
