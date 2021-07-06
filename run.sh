# remove prior containers
for image in `docker image ls | grep none | cut -d'>' -f3 | cut -d' ' -f8`; do echo "For outdated image $image:"; for container in `docker ps -a | grep $image | cut -d' ' -f1`; do echo "  Remove container $container"; docker stop $container; docker rm $container; done; docker image rm $image; done;
for container in `docker ps | grep go-docker | grep latest | cut -d' ' -f1`; do echo "Remove old but current container $container"; docker stop $container; docker rm $container; done;

echo "Run new container"
docker run -d \
--read-only \
--network host \
-e LOG_FILE_LOCATION=/logs/go-docker.main.log \
-e LOG_FILE_ROOT=/logs/ \
-v /docker-vols/go-docker-logs/:/logs/ \
-v /docker-vols/go-docker-store/:/store/ \
go-docker:latest
echo $?

export CONTAINER=`docker ps -a | grep go-docker:latest | cut -d' ' -f1`
for container in $CONTAINER; do echo docker exec -it $container bash; echo docker logs $container; done

# docker exec -it $CONTAINER bash
docker ps -a | grep go-docker
echo tail -f /docker-vols/go-docker-logs/go-docker.main.log
tail -f /docker-vols/go-docker-logs/go-docker.main.log