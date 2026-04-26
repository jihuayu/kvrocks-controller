
# HTTP APIs

When `auth.type` is set to `local`, existing resource APIs require a JWT bearer token:

```http
Authorization: Bearer <token>
```

## Authentication APIs

### Login

```shell
POST /api/v1/auth/login
```

#### Request Body

```json
{
  "username": "admin",
  "password": "admin-password"
}
```

#### Response JSON Body

* 200
```json
{
  "data": {
    "token": "JWT_TOKEN",
    "expires_at": "2026-04-26T12:00:00Z",
    "user": {
      "username": "admin",
      "role": "admin",
      "created_at": "2026-04-26T11:00:00Z",
      "updated_at": "2026-04-26T11:00:00Z"
    }
  }
}
```

* 401
```json
{
  "error": {
    "message": "unauthorized"
  }
}
```

### Get Current User

```shell
GET /api/v1/auth/me
```

#### Response JSON Body

* 200
```json
{
  "data": {
    "auth_enabled": true,
    "user": {
      "username": "admin",
      "role": "admin"
    }
  }
}
```

JWTs are stateless and are not revoked by the controller. Clients should discard the token locally when signing out.

## User APIs

User APIs are available only when local authentication is enabled, and only administrators can call them.

### Create User

```shell
POST /api/v1/user
```

#### Request Body

```json
{
  "username": "dev",
  "password": "dev-password",
  "role": "user"
}
```

#### Response JSON Body

* 201
```json
{
  "data": {
    "user": {
      "username": "dev",
      "role": "user",
      "created_at": "2026-04-26T11:00:00Z",
      "updated_at": "2026-04-26T11:00:00Z"
    }
  }
}
```

### List Users

```shell
GET /api/v1/user
```

#### Response JSON Body

* 200
```json
{
  "data": {
    "users": [
      {
        "username": "admin",
        "role": "admin",
        "created_at": "2026-04-26T11:00:00Z",
        "updated_at": "2026-04-26T11:00:00Z"
      }
    ]
  }
}
```

### Get User

```shell
GET /api/v1/user/{username}
```

### Update User

```shell
PUT /api/v1/user/{username}
```

#### Request Body

```json
{
  "password": "new-password",
  "role": "admin"
}
```

### Delete User

```shell
DELETE /api/v1/user/{username}
```

The last administrator cannot be deleted or demoted.

## Namespace APIs
### Create Namespace

```shell
POST /api/v1/namespaces
```

#### Request Body

```json
{
  "namespace": "test-ns"
}
```

#### Response JSON Body

* 201
```json
{
  "data": "created"
}
```

* 409
```json
{
  "error": {
    "message": "the entry already existed"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

### List Namespace
```shell
GET /api/v1/namespaces
```
#### Response JSON Body

* 200
```json
{
  "data": {
    "namespaces": ["test-ns"]
  }
}
```
* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

### Delete Namespace

```shell
GET /api/v1/namespaces/{namespace}
```

#### Response JSON Body

* 200
```json
{
  "data": "ok"
}
```

* 404
```json
{
  "error": {
    "message": "the entry does not exist"
  }
}

```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```
## Cluster APIs
### Create Cluster

```
POST /api/v1/namespaces/{namespace}/clusters
```

#### Request Body

```json
{
  "name":"test-cluster",
  "nodes":["127.0.0.1:6666"],
  "replicas":1,
  "password":""
}
```

#### Response JSON Body

* 201
```json
{
  "data": "created"
}
```

* 409
```json
{
  "error": {
    "message": "the entry already existed"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

### List Cluster

```shell
GET /api/v1/namespaces/{namespace}/clusters
```
#### Response JSON Body

* 200
```json
{
  "data": {
    "clusters": ["test-cluster"]
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

### Import Cluster

This API is used to import the cluster from the existing Kvrocks cluster's nodes.

```shell
POST /api/v1/namespaces/{namespace}/clusters/{cluster}/import
```

#### Request Body

```json
{
  "nodes":["127.0.0.1:6666"],
  "password":""
}
```

#### Response JSON Body

* 201
```json
{
  "data": "created"
}
```

* 409
```json
{
  "error": {
    "message": "the entry already existed"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

### Get Cluster

```shell
GET /api/v1/namespaces/{namespace}/clusters/{cluster}
```

#### Response JSON Body

* 200
```json
{
  "data":
  {
    "cluster": {
      "name":"test-cluster",
      "version":0,
      "shards":[
        {"nodes":[
          {
            "id":"YotDSqzTeHK6CnIX2gZu27IlcYRTW4dkkFQvV382",
            "addr":"127.0.0.1:6666",
            "role":"master",
            "password":"",
            "master_auth":"",
            "created_at":16834433980
          }],
          "slot_ranges":["0-16383"],
          "import_slot":-1,
          "migrating_slot":-1
        }
      ]
    }
  }
}
```

* 404
```json
{
  "error": {
    "message": "the entry does not exist"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

### Delete Cluster

```shell
DELETE /api/v1/namespaces/{namespace}/clusters/{cluster}
```

#### Response JSON Body

* 200
```json
{
  "data": "ok"
}
```

* 404
```json
{
  "error": {
    "message": "the entry does not exist"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```
## Shard APIs
### Create Shard 

```
POST /api/v1/namespaces/{namespace}/clusters/{cluster}/shards
```

#### Request Body

```json
{
  "nodes":["127.0.0.1:6666"],
  "password":""
}
```

#### Response JSON Body

* 201
```json
{
  "data": "created"
}
```

* 409
```json
{
  "error": {
    "message": "the entry already existed"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```


### Get Shard

```shell
GET /api/v1/namespaces/{namespace}/clusters/{cluster}/shards/{shard}
```

#### Response JSON Body

* 200
```json
{
  "data": {
    "shard": {
      "nodes": [
        {
          "id": "3SStZULMqclwvYNT8gN05IdybROe0vEnn97iNB5Z",
          "addr": "127.0.0.1:6666",
          "role": "master",
          "password": "",
          "master_auth": "",
          "created_at": 16834433980
        }
      ],
      "slot_ranges": [
        "0-16383"
      ],
      "import_slot": -1,
      "migrating_slot": -1
    }
  }
}

```

* 404
```json
{
  "error": {
    "message": "the entry does not exist"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

### List Shard 

```shell
GET /api/v1/namespaces/{namespace}/clusters/{cluster}/shards
```
#### Response JSON Body

* 200
```json
{
  "data": {
    "shards": [
      {
        "nodes": [
          {
            "id": "3SStZULMqclwvYNT8gN05IdybROe0vEnn97iNB5Z",
            "addr": "127.0.0.1:6666",
            "role": "master",
            "password": "",
            "master_auth": "",
            "created_at": 16834433980
          }
        ],
        "slot_ranges": [
          "0-16383"
        ],
        "import_slot": -1,
        "migrating_slot": -1
      },
      {
        "nodes": [
          {
            "id": "y5PftTd0Lc3hH34yEyavIji86cRM5i3oxytt42vo",
            "addr": "127.0.0.1:6667",
            "role": "master",
            "password": "",
            "master_auth": "",
            "created_at": 16834433980
          }
        ],
        "slot_ranges": null,
        "import_slot": -1,
        "migrating_slot": -1
      }
    ]
  }
}

```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

### Delete Shard 

```shell
DELETE /api/v1/namespaces/{namespace}/clusters/{cluster}/shards/{shard}
```

#### Response JSON Body

* 200
```json
{
  "data": "ok"
}
```

* 404
```json
{
  "error": {
    "message": "the entry does not exist"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

### Failover master node in a shard

```
POST /api/v1/namespaces/{namespace}/clusters/{cluster}/shards/{shard}/failover
```

#### Request Body

```json
{
  "preferred_node_id": "{YOUR PREFERRED NODE ID}"
}
```

#### Response JSON Body

* 200
```json
{
  "data": {
    "new_master_id": "{NEW MASTER ID}"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

## Node APIs

### Create Node

```
POST /api/v1/namespaces/{namespace}/clusters/{cluster}/shards/{shard}/nodes
```

#### Request Body

```json
{
  "addr": "127.0.0.1:6666",
  "role": "slave",
  "password":""
}
```

#### Response JSON Body

* 201
```json
{
  "data": "created"
}
```

* 409
```json
{
  "error": {
    "message": "the entry already existed"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

### List Node 

```shell
GET /api/v1/namespaces/{namespace}/clusters/{cluster}/shards/{shard}/nodes
```
#### Response JSON Body

* 200
```json
{
  "data": {
    "nodes": [
      {
        "id": "pxpE1JSBJcqicuwc95zqTPTj5rB7YtfvpociyH8C",
        "addr": "127.0.0.1:6666",
        "role": "master",
        "password": "",
        "created_at": 1686101693
      },
      {
        "id": "O0JKq1Hp9FtI3dJTU3MigWjjZJzPtduoDODX0OAY",
        "addr": "127.0.0.1:6667",
        "role": "slave",
        "password": "",
        "created_at": 1686102057
      }
    ]
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

### Delete Node 

```shell
DELETE /api/v1/namespaces/{namespace}/clusters/{cluster}/shards/{shard}/nodes/{nodeID}
```

#### Response JSON Body

* 200
```json
{
  "data": "ok"
}
```

* 404
```json
{
  "error": {
    "message": "the entry does not exist"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```

## Migration APIs

### Migrate Slot
```shell
POST /api/v1/namespaces/{namespace}/clusters/{cluster}/migrate
```

#### Request Body

```json
{
  "target": 1,
  "slot": 123,
  "slot_only": "false" 
}
```

#### Response JSON Body

* 200
```json
{
  "data": "ok"
}
```

* 404
```json
{
  "error": {
    "message": "the entry does not exist"
  }
}
```

* 5XX
```json
{
  "error": {
    "message": "DETAIL ERROR STRING"
  }
}
```
