### Icecream
[![GoDoc](https://godoc.org/github.com/cfchou/icecream?status.svg)](https://godoc.org/github.com/cfchou/icecream)

An API server.

### Requirement:
- go: My development is done on go 1.11. However, 1.8+ should work.
- docker: For running the database.
- make 
- curl: It is needed if you want to run the examples.

### Installation

- Make sure go and __$GOPATH__ are set up properly and docker is running.
- Run these commands:
    ```
    mkdir -p $GOPATH/src/github.com/cfchou
    cd $GOPATH/src/github.com/cfchou
    git clone https://github.com/cfchou/icecream.git
    cd icecream
    make db
    make test
    make run
    ```
Now we should be able to see the API server running in foreground.


### Configuration

A default configuration _icecream.yaml_ is provided. If you want to serve apiserver with SSL, then you have to provide the location of the certificate and the key in the config. 

For finer control, a _Makefile_ is provided:
- make test: run unit test.
- make apiserver: build the binary 
- make db: recreate the db and preload data.
- make run


### Design

##### Database
As to the database for this project, I choose mongoDB, a schemaless document-oriented database which stores JSON natively. There are a few immediate advantages. Firstly, a sample from the dataset enclosed in this project is a document in JSON, so it's easy to load the data into the db. Secondly, I decide to manipulate a document in a whole rather than break it down to columns in different tables with foreign keys pointing to each other. RDBMS has the merit of maintaining strong consistency of data and detecting violations for us. However, for this small project, I would favor mongoDB's simplicity over RDBMS' capability.

Nevertheless, there are a few caveats of using mongoDB here, the biggest problem for me is the lack of SQL to manipulate the data. Also constraints like uniqueness of fields have to be done by issuing specific mongoDB commands.


##### Server Design
I don't make use of a web framework as this is a simple RESTful server. Having said that, I do rely on some 3rd-party libraries to build this project. Just to name a few, [spf13/viper](https://github.com/spf13/viper) for configuration, [gorilla/mux](http://www.gorillatoolkit.org/pkg/mux) for URL routing, [inconshreveable/log15](https://github.com/inconshreveable/log15) for contextual logging, and [stretchr/testify](https://github.com/stretchr/testify) for testing and mocking.

The http-related source code sits in the __cmd/apiserver__. This service can be extended by adding __middleware__ to support less business related operations like auditing, metrics, etc.. At the moment, there's only one and it's for API key authentication. The real CRUD logic is implemented in __handler__ which connects to backend of choice to access the data.

I try to make data access layer and models reusable and extensible. As the result, it is implemented as a backend in __pkg/backend__. I have done one for mongoDB. But it's possible to write others for redis, RDBMS, and even cloud storages.


### API

##### Authentication
For authentication/authorization, __Authorization: $YOUR_API_KEYS__ has to be in the HTTP header. This design is common and efficient, but it has to go with SSL to be secure. There are two API keys preloaded in the db for test:

- "testkey"
- "0123456789"


##### API list
Looking into the sample data, I assume each icecream product is uniquely identified by the field __productId__.
The goal is to support CRUD for products. APIs are listed below:


###### Create:
* POST /products/
    
This is an __exclusive create__. It fails if there's already a s product in the db with the same productId.
```
cat <<HERE | curl -i -XPOST --header "Authorization: testkey" localhost:8080/products/ -d @-
{
    "allergy_info": "",
    "description": "",
    "dietary_certifications": "",
    "image_closed": "",
    "image_open": "",
    "ingredients": [],
    "name": "Test1",
    "productId": "001",
    "sourcing_values": [],
    "story": ""
}
HERE
```

* PUT /products/{productID}
    
This is an __upsert__. It creates the product if not existed, otherwise it replaces what's in the db.
```
cat <<HERE | curl -i -XPUT --header "Authorization: testkey" localhost:8080/products/001 -d @-
{
    "allergy_info": "",
    "description": "updattttttttttte",
    "dietary_certifications": "",
    "image_closed": "",
    "image_open": "",
    "ingredients": [],
    "name": "Test1",
    "productId": "001",
    "sourcing_values": [],
    "story": ""
}
HERE
```

* Payload for APIs above are expected to be __all fields__ of a product in json.


###### Update:
* PUT /products/{productID}

Fully update. It is the same as the one for Create.


* PATCH /products/{productID}

Partial update.
```
cat <<HERE | curl -i -XPATCH --header  "Authorization: testkey" localhost:8080/products/001 -d @-
{
    "description": "extra information",
    "ingredients": ["soy", "milk"]
}
HERE
```


###### Read:
* GET /products/{productID}

Read a product.
```
curl -i -XGET --header "Authorization: testkey" localhost:8080/products/2188
```

* Get /products/\[?__cursor=$cursor__&__limit=$limit__\]

Read products(support pagination). __$cursor__ is the end of last page, __$limit__ is the max number of products per page. $limit is capped by _limitToRead_ in _icecream.yaml_. It returns products and the next __$cursor__.
```
curl -i -XGET --header "Authorization: testkey" localhost:8080/products/
curl -i -XGET --header "Authorization: testkey" localhost:8080/products/\?limit=2
curl -i -XGET --header "Authorization: testkey" localhost:8080/products/\?limit=2\&cursor=5bbaeea1246ed82dc66b2603
```


###### Delete:
* DELETE /products/{productID}

Delete a product.
```
curl -i -XDELETE --header "Authorization: testkey" localhost:8080/products/001
```


### TODO
- Higher test coverage
- Improve error handling
- GraphQL


