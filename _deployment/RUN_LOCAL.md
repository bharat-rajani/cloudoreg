Run postgres, redis, and kafka with net=host

docker run -d --net=host --name sources-postgres \
    -e POSTGRES_PASSWORD=mysecretpassword \
    -e PGDATA=/var/lib/postgresql/data/pgdata \
    -v $(pwd)/deployment/data/postgres:/var/lib/postgresql/data postgres

docker run --name sources-redis --net=host -d redis

docker run --rm -it --net=host --name=kafka -e RUNTESTS=0 --pull=always lensesio/fast-data-dev


Clone sources-api-go
git clone git@github.com:RedHatInsights/sources-api-go.git

change directory to cloned sources-api-go and run:
source /YOUR_CLONE_DIR/cloudoreg/deployment/env.sh
make setup
make inlinerun


Create a json file for /bulk_create body
```json
{
  "sources": [
    {
      "name": "bharat_test_source",
      "app_creation_workflow": "manual_configuration",
      "source_type_name": "amazon"
    }
  ],
  "endpoints": [],
  "authentications": [
    {
      "authtype": "cloud-meter-arn",
      "username": "arn:aws:iam::665427542893:role/TestBRRole",
      "resource_type": "application",
      "resource_name": "/insights/platform/cloud-meter"
    }
  ],
  "applications": [
    {
      "application_type_id": "3",
      "extra": {
        "hcs": false
      },
      "source_name": "bharat_test_source"
    }
  ]
}

```
curl -XPOST localhost:4000/api/sources/v3.1/bulk_create -d @/tmp/sourceBody.json -H "x-rh-sources-account-number: 1234466"