## API Docs (WIP)
This document describes the APIs exposed by the **Config Server**.

### 1 - Get By ID
```
GET /v1/data/:id
```

| Id  | Description |
| ---- | ----------- |
| id | string unique id   |

##### Response Body
`Content-Type: application/json`

| Id| Type | Description |
| ----| ---- | ---- | ----------- |
| id | string | Unique Id |
| name | string | Full path |
| value | JSON Object | Any valid JSON object |

##### Response Codes
| Code   | Description |
| ------ | ----------- |
| 200 | Status OK |
| 400 | Bad Request |
| 401 | Not Authorized |
| 404 | Name not found |
| 500 | Server Error |

##### Sample Requests/Responses

`GET /v1/data/:id`

Response:
``` JSON
{
  "id": "some_id",
  "name": "color",
  "value": "blue"
}
```

--

### 2 - Get By Name

`GET /v1/data?name="/server/tomcat/port"`

Response:
``` JSON
{
  "data": [
    {
      "id": "some_id",
      "name": "server/tomcat/port",
      "value": 8080
    }
  ]

```

--

`GET /v1/data?name="/server/tomcat/cert"`

Response:
``` JSON
{
  "data": [
    {
      "id": "some_id",
      "name": "server/tomcat/port",
      "value": {
        "cert": "my-cert",
        "private-key": "my private key"
      }
    }
  ]
}
```

---

### 3 - Set Name Value
```
PUT /v1/data
```

##### Request Body
`Content-Type: application/json`

| Name | Type | Description |
| ---- | ---- | ----------- |
| name| string | name of key | 
| value | JSON Object | Any valid JSON object |

##### Sample Request

`PUT /v1/data`

Request Body:
```
{
  "name": "full/path/to/name",
  "value": "happy value"
}
```

##### Response Body
`Content-Type: application/json`

| Name | Type | Description |
| ---- | ---- | ----------- |
| id | string | Unique Id |
| name | string | Full path |
| value | JSON Object | Any valid JSON object |

##### Response Codes
| Code | Description |
| ---- | ----------- |
| 200 | Call successful - name value was added |
| 400 | Bad Request |
| 401 | Not Authorized |
| 415 | Unsupported Media Type |
| 500 | Server Error |

---

### 4 - Generate password/certificate

```
POST /v1/data/
```

##### Request Body
`Content-Type: application/json`

| Name | Type | Valid Values | Description |
| ---- | ---- | ------------ | ----------- |
| name | String | alphanumeric | name of key |
| type | String | password, certificate | The type of data to generate |
| parameters | JSON Object | | See below for valid parameters |

###### Request body extra parameters values
| For type | Name | Type |
| -------- | ---- | ---- |
| certificate | common_name | String |
| certificate | alternative_names | Array of Strings |

##### Sample Requests

###### Password Generation
`POST /v1/data`

Request Body:
``` JSON
{
  "name": "mypasswd",
  "type": "password"
}
```

###### Certificate Generation
`POST /v1/data`

Request Body:
``` JSON
{
  "name": "mycert",
  "type": "certificate",
  "parameters": {
    "common_name": "bosh.io",
    "alternative_names": ["bosh.io", "blah.bosh.io"]
  }
}
```

##### Response Body
`Content-Type: application/json`

It returns an array of the following object:

| Name | Type | Description |
| ---- | ---- | ----------- |
| name | string | Name of key |
| id | string | Unique Id |
| name | string | Full path  |
| value | JSON Object | value generated |

##### Response Codes
| Code | Description |
| ---- | ----------- |
| 201 | Call successful |
| 400 | Bad Request |
| 401 | Not Authorized |
| 415 | Unsupported Media Type |
| 500 | Server Error |

##### Sample Response
###### Password
```
{
  "id": "some_id",
  "name": "/mypasswd",
  "value":"49cek4ow75ev5zw4t3v3"
}
```
###### Certificate
``` 
{
  "id": "some_id",
  "name":"/mycert",
  "value": {
    "ca" : "---- Root CA Certificate ----",
    "certificate": "---- Generated Certificate. Signed by rootCA ----",
    "private_key": "---- Private key for the Generated certificate ----"
  }
}
```

### 5 - Delete Name
```
DELETE /v1/data?name="name"
```

| Name | Description |
| ---- | ----------- |
| name | Full path |


##### Sample Request

`DELETE /v1/data?name="full/path/to/name"`

##### Response Codes
| Code | Description |
| ---- | ----------- |
| 204 | Call successful - name was deleted |
| 400 | Bad Request |
| 401 | Not Authorized |
| 404 | Not Found |
| 500 | Server Error |
