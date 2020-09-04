# Filter API POC
POC app demoing integration of the Filter API with Flexible Table Builder (FTB).

## Overview
When querying a _Flexible dataset_ the result set is run through a set of _disclosure control rules_ (DCR). These rules ensure 
that an individual cannot be identified from the resulting observation values. When filtering a flexible dataset the 
filter API needs to execute a query for the current set of options and return the DCR status to the user. This will allow
 a user to build up their filter receiving immediate feedback on the validity of their option choices.
 
The POC aims to demonstrate that its possible to build this functionality into the existing Filter. Specifically:
- Can we build the necessary FTB query using the existing Filter API url structure.
- Can we map the FTB query response data into the existing Filter response structure.
- How we might present the DCR status of a query to users.  

The app implements the same interface as the real Filter API but internally queries the FTB instead of following the CMD
 filter process. The `GET Filter` response entity has been extended to include a new optional field to store the 
 disclosure control status details.

## Config
Set the following env vars:

| Env var name       | Description                                                       |
|:-------------------|:------------------------------------------------------------------|
| `EC2_IP`           | The IP address of the FTB API running on the develop AWS account. |
| `AUTH_PROXY_TOKEN` | The auth token value required by the FTB instance                 |

## Run
```
make debug
```
The app runs on port `22100` by default

### Example
**Note**: Example queries assume the FTB instance has loaded the `People` dataset.

Get the filter job.
```
curl "http://localhost:22100/filters/12345" | jq
```

Adding dimensions
```bash
curl -XPOST "http://localhost:22100/filters/12345/dimensions/sex/options/1" | jq
```

```
curl -XPOST "http://localhost:22100/filters/12345/dimensions/age/options/31" | jq
```

Get the filter and check the status filter
```
curl "http://localhost:22100/filters/12345" | jq
```

```json
{
  "dataset": {
    "id": "People",
    "edition": "time-series",
    "version": 1
  },
  "instance_id": "",
  "dimensions": [
    {
      "name": "OA",
      "options": []
    },
    {
      "name": "Age",
      "options": [
        "31"
      ]
    },
    {
      "name": "Sex",
      "options": [
        "1"
      ]
    }
  ],
  "filter_id": "12345",
  "links": {
    "dimensions": {},
    "filter_output": {},
    "filter_blueprint": {},
    "self": {},
    "version": {}
  },
  "disclosure_control": {
    "status": "OK",
    "dimension": "OA",
    "options": [],
    "count": 0
  }
}
```

Adding an output area option will trigger the DC rules and blocks the response. 
```
curl -XPOST "http://localhost:22100/filters/12345/dimensions/oa/options/synW00000005" | jq
```

```json
{
  "dataset": {
    "id": "People",
    "edition": "time-series",
    "version": 1
  },
  "instance_id": "",
  "dimensions": [
    {
      "name": "OA",
      "options": [
        "synW00000005"
      ]
    },
    {
      "name": "Age",
      "options": [
        "31"
      ]
    },
    {
      "name": "Sex",
      "options": [
        "1"
      ]
    }
  ],
  "filter_id": "12345",
  "links": {
    "dimensions": {},
    "filter_output": {},
    "filter_blueprint": {},
    "self": {},
    "version": {}
  },
  "disclosure_control": {
    "status": "Blocked",
    "dimension": "OA",
    "options": [],
    "count": 1
  }
}
```

The exact format of the disclosure status and what data should be included is up for discussion. This is simply for illustration purposes. 