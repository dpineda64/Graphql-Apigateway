# GRAPHQL API GATEWAY

Search for microservices and build a dynamically generated graphql Schema
based on service Request and Response

Example
```
syntax = "proto3";

package go.graphql.test;

service Users {
    rpc Get(Request) returns (User) {}
}

message Request {
    int32 id = 1;
}

message User {
    bool done =1;
    int32 id = 2;
    string username = 3;
}
```

that is transformed into a Graphql query like this
```
Users_Get(id: Int): Users_Get
```



_**NOTE** this does not convert from proto to graphql syntax_

_Not tested on production or with complex queries and mutations are not implemented yet_

Build with
- [micro](https://github.com/micro/go-micro)
- [graphql-go](https://github.com/graphql-go/graphql)
- [kubernetes](https://kubernetes.io)
- [rabbitmq](http://www.rabbitmq.com/)
- [gnatsd]( https://github.com/nats-io/gnatsd)
 
