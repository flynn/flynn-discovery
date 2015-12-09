# Flynn Discovery

A cluster peer discovery server.

`flynn-discovery` provides a simple HTTP interface to create clusters, register cluster members and get cluster members.

To get `flynn-discovery` run (requires a correctly configured Go environment):

```
$ go get github.com/flynn/flynn-discovery
```

To deploy `flynn-discovery` into a Flynn cluster execute the following steps:

```
$ cd $GOPATH/src/github.com/flynn/flynn-discovery
$ flynn create flynn-discovery
$ flynn resource add postgres
$ cat schema.sql | flynn pg psql --
$ git push flynn master
```

At this point you should have a deployed version of `flynn-discovery` running on your Flynn cluster.

To see the address run `flynn route`:

```
$ flynn route
ROUTE                                    SERVICE              ID                                         STICKY
http:flynn-discovery.dev.localflynn.com  flynn-discovery-web  http/5bb8495e-a660-430d-baa3-4995972ab020  false
```

Export that URL to use in tests:

```
$ export FLYNN_DISCOVERY_URL=$(flynn route | sed 1d | cut -f 1 -d " " | sed 's/^http://')
```

Now we can test `flynn-discovery`.

First we need to create a cluster token. This will uniquely identify the cluster. The normal workflow is to create the cluster token beforehand and give it to each cluster member so they can register themselves as members of that cluster.

```
$ curl -XPOST $FLYNN_DISCOVERY_URL/clusters -I
HTTP/1.1 201 Created
Location: /clusters/e99a6a09-bc2b-4dbb-b84e-c70ae176be48
Date: Thu, 26 Nov 2015 12:26:54 GMT
Content-Length: 0
Content-Type: text/plain; charset=utf-8
```

If the token was created successfully we should get a `201 Created` response status. The `Location` header is the (relative) URL representing the cluster token.

The cluster token is `http://$FLYNN_DISCOVERY_URL/clusters/e99a6a09-bc2b-4dbb-b84e-c70ae176be48`.

Next we can add cluster members by using the provided cluster token:

```
$ curl -XPOST $FLYNN_DISCOVERY_URL/clusters/e99a6a09-bc2b-4dbb-b84e-c70ae176be48/instances -d '
{
  "data": {
    "name": "instance-2",
    "url": "http://localhost:3333"
  }
}
'
{"data":{"id":"66b52ea9-50b9-41ab-842e-72643b833400","cluster_id":"e99a6a09-bc2b-4dbb-b84e-c70ae176be48","url":"http://localhost:2222","name":"instance-1","created_at":"2015-11-26T12:24:32.580008Z"}}v
```

Add another cluster member:

```
$ curl -XPOST $FLYNN_DISCOVERY_URL/clusters/e99a6a09-bc2b-4dbb-b84e-c70ae176be48/instances -d '
> {
>   "data": {
>     "name": "instance-2",
>     "url": "http://localhost:3333"
>   }
> }
> '
{"data":{"id":"b4b84b79-4ba5-4847-9827-41616e0db056","cluster_id":"e99a6a09-bc2b-4dbb-b84e-c70ae176be48","url":"http://localhost:3333","name":"instance-2","created_at":"2015-11-26T12:25:25.745206Z"}}
```

List all the cluster members:

```
$ curl $FLYNN_DISCOVERY_URL/clusters/e99a6a09-bc2b-4dbb-b84e-c70ae176be48/instances
{"data":[{"id":"66b52ea9-50b9-41ab-842e-72643b833400","cluster_id":"e99a6a09-bc2b-4dbb-b84e-c70ae176be48","url":"http://localhost:2222","name":"instance-1","created_at":"2015-11-26T12:24:32.580008Z"},{"id":"b4b84b79-4ba5-4847-9827-41616e0db056","cluster_id":"e99a6a09-bc2b-4dbb-b84e-c70ae176be48","url":"http://localhost:3333","name":"instance-2","created_at":"2015-11-26T12:25:25.745206Z"}]}
```

Or pretty printed:

```
$ curl -sS $FLYNN_DISCOVERY_URL/clusters/e99a6a09-bc2b-4dbb-b84e-c70ae176be48/instances | json_pp
{
   "data" : [
      {
         "url" : "http://localhost:2222",
         "id" : "66b52ea9-50b9-41ab-842e-72643b833400",
         "name" : "instance-1",
         "cluster_id" : "e99a6a09-bc2b-4dbb-b84e-c70ae176be48",
         "created_at" : "2015-11-26T12:24:32.580008Z"
      },
      {
         "cluster_id" : "e99a6a09-bc2b-4dbb-b84e-c70ae176be48",
         "created_at" : "2015-11-26T12:25:25.745206Z",
         "url" : "http://localhost:3333",
         "name" : "instance-2",
         "id" : "b4b84b79-4ba5-4847-9827-41616e0db056"
      }
   ]
}
```
