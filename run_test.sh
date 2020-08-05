#!/bin/bash

# This script DOESN'T remove the container when completed to avoid needing to
# setup a brand new database for every run, as the first initialization takes a
# long time.

# This script requires:
# - docker
# - gotest: https://github.com/rakyll/gotest
# - mysql client
# - gcc (for race detection, recommended but optional)

echo -e "\033[1mStarting database containers...\033[0m"
if [ ! "$(docker ps -a | grep goyave-mariadb)" ]; then
	container=`docker run --name goyave-mariadb -p 3306:3306 -e MYSQL_ROOT_PASSWORD=secret -e MYSQL_USER=goyave -e MYSQL_PASSWORD=secret -e MYSQL_DATABASE=goyave -d mariadb:latest`
else
	container=`docker start goyave-mariadb`
fi

if [ $? -ne 0 ]; then
	echo -e "\033[31mError: couldn't start database container.\033[0m"
	exit $?
fi

health=1
tries=0
while [ $health -ne 0 ]; do
	mysql -h 127.0.0.1 --protocol=tcp -u goyave --password=secret goyave -e "SHOW TABLES;" >/dev/null 2>&1
	health=$?
	((tries++))
	if [ $tries -gt 100 ]; then
		docker stop $container
		echo -e "\033[31mError: couldn't connect to container database after 100 retries.\033[0m"
		exit 2
	fi
	echo -e "\033[33mCouldn't connect to container database. Retrying in 5 seconds...\033[0m"
	sleep 5
done

echo -e "\033[92m\033[1mDatabase ready. Running tests...\033[0m"
gcc --version >/dev/null 2>&1
if [ $? -ne 0 ]; then
	echo -e "\033[33mgcc is missing. Running tests without data race checking.\033[0m"
	gotest -v -p 1 -coverprofile=c.out -coverpkg=./... ./... ; go tool cover -html=c.out ; go tool cover -func=c.out | grep total ; rm c.out
else
	gotest -v -p 1 -race -coverprofile=c.out -coverpkg=./... ./... ; go tool cover -html=c.out ; go tool cover -func=c.out | grep total ; rm c.out
fi
test_result=$?

echo -e "\033[1mStopping database container...\033[0m"
docker stop $container >/dev/null

exit $test_result