#!/usr/bin/env bash -e

CONTAINER=my_mongo

# Location where data of mongoDB persist
ICECREAM_JSON=/Users/cfchou/Project/gopath/src/bitbucket.org/cfchou/icecream/icecream.json
APIKEY_JSON=/Users/cfchou/Project/gopath/src/bitbucket.org/cfchou/icecream/apikey.json
#
DB=icecream

# aliases
mongod="docker run --name $CONTAINER \
           --rm -p 27017:27017 \
           -d mongo:3.6"
mongod_stop="docker stop $CONTAINER"
mongo_import="docker exec -i my_mongo mongoimport --db $DB --drop --collection "

## icecream.json is a json array, reformat it to a stream of json objects.
#jq="docker run -i --rm stedolan/jq '.[]'"


#
echo "(Re)Start monogoDB......"
$mongod_stop || echo ""
$mongod

echo "Reload data from $ICECREAM_JSON......"
cat $ICECREAM_JSON | $(echo "$mongo_import product")
echo "Reload data from $APIKEY_JSON......"
cat $APIKEY_JSON | $(echo "$mongo_import apikey")









