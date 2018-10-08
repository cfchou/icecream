#!/usr/bin/env bash -e

CONTAINER=my_mongo

# data preload to mongoDB
ICECREAM_JSON=icecream.json
APIKEY_JSON=apikey.json
DB=icecream

# aliases
mongod="docker run --name $CONTAINER \
           --rm -p 27017:27017 \
           -d mongo:3.6"
mongod_stop="docker stop $CONTAINER"
mongo_import="docker exec -i my_mongo mongoimport --db $DB --drop --collection "


#
echo "(Re)Start monogoDB......"
$mongod_stop 2>/dev/null || echo ""
$mongod

echo "Reload data from $ICECREAM_JSON......"
cat $ICECREAM_JSON | $(echo "$mongo_import products")
echo "Reload data from $APIKEY_JSON......"
cat $APIKEY_JSON | $(echo "$mongo_import apikeys")









