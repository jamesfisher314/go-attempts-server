# remove prior containers
for image in `docker image ls | grep none | cut -d'>' -f3 | cut -d' ' -f5`; do echo "For outdated image $image:"; for container in `docker ps -a | grep $image | cut -d' ' -f1`; do echo "  Remove container $container"; docker stop $container; docker rm $container; done; docker image rm $image; done;
for container in `docker ps | grep go-docker | grep latest | cut -d' ' -f1`; do echo "Remove old but current container $container"; docker stop $container; docker rm $container; done;
echo "Run new container"
export CONTAINER=`docker run -d \
--read-only \
-p 8080:8080 \
-e LOG_FILE_LOCATION=/logs/go-docker.main.log \
-v /docker-vols/go-docker-logs/:/logs/ \
-v /docker-vols/go-docker-store/:/store/ \
go-docker:latest`
echo docker exec -it $CONTAINER bash
docker exec -it $CONTAINER bash