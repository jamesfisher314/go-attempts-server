docker image ls | grep none | cut -d'>' -f3 | cut -d' ' -f8 | xargs -n1 docker image rm
docker ps -a | grep -v IMAGE | cut -d' ' -f1 | xargs -n1 docker rm

