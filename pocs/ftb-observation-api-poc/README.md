# FTB Observation API POC

POC app demoing integration of FTB queries into the observation API. The app uses the URL structures 
as the real observation api but internally it queries FTB in instead of Neo4j/Neptune. Returns the query result in the 
observation API response structure.

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
The app runs on port `24500` by default

### Queries
**Note**:
 - Example queries assume the FTB instance has loaded the `People` dataset.
 - A wildcard is specified by a parameter with a name and no value e.g. `&sex` instead of `&sex=*`.

Query for a single observation:
```
curl "http://localhost:24500/datasets/People/editions/time-series/versions/1/observations?country=synE92000001&age=31&sex=1" | jq .
```
Query with a wild card for a single dimension
```
curl "http://localhost:24500/datasets/People/editions/time-series/versions/1/observations?country=synE92000001&age=31&sex" | jq .
```

Query with multiple options selected for a dimension
```
curl "http://localhost:24500/datasets/People/editions/time-series/versions/1/observations?country=synE92000001&age=31&age=30&sex" | jq .
```

Query with multiple wildcards
```
curl "http://localhost:24500/datasets/People/editions/time-series/versions/1/observations?country=synE92000001&age&sex" | jq .
```
