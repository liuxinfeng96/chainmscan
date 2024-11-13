#!/bin/bash
path=`pwd`
version=$1

i=$(docker images | grep "chainmscan-server" | grep "$version" | awk '{print $1}')
if test -z $i; then
echo "not found the docker image, start build image..."
docker build -f ./DockerFile -t chainmscan-server:$version ../chainmscan
fi

i=$(docker images | grep "chainmscan-server" | grep "$version" | awk '{print $1}')
if test -z $i; then
echo "build image error, exit shell!"
exit
fi

c=$(docker ps -a | grep "chainmscan-mysql-$version" | awk '{print $1}')
if test -z $c; then
echo "not found the mysql server, start mysql server..."

docker run -d \
    -p 33065:3306 \
    -v $path/conf/my.cnf:/etc/mysql/mysql.conf.d/my.cnf \
    -v $path/../chainmscan-data:/var/lib/mysql \
    -e MYSQL_ROOT_PASSWORD=123456 \
    -e MYSQL_DATABASE=chainmscan \
    --name chainmscan-mysql-$version \
    --restart always \
    mysql:8.0
echo "waiting for database initialization..."
sleep 20s
docker logs --tail=10 chainmscan-mysql-$version
fi

i=$(docker ps -a | grep "chainmscan-server:$version" | awk '{print $1}')
if test ! -z $i; then
echo "the server container already exists, delete..."
docker rm -f chainmscan-server-$version
fi

echo "start the server..."
docker run -d \
    -p 9660:9660 \
    -w /chainmscan \
    -v $path/conf:/chainmscan/conf \
    -v $path/log:/chainmscan/log \
    -v $path/tmp:/chainmscan/tmp \
    -e TZ=Asia/Shanghai \
    -m 1024M \
    --net=host \
    --memory-swap 2048M \
    --cpus 2 \
    --name chainmscan-server-$version \
    --restart always \
    --privileged \
    chainmscan-server:$version \
    bash -c "./chainmscan -config ./conf/config.yaml"
sleep 2s
docker logs chainmscan-server-$version
echo "the server has been started!"

