To run locally:  
1. Source the required configs: `source _deployment/env.sh`
2. Execute: `make run`  
This will run `sources-api` along with all its dependencies. This is a blocking call.
3. In a separate terminal, execute: ```make bulk_create```  
This will send the http request to sources-api.