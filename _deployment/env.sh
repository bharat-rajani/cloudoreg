# Kafka related setup
export QUEUE_HOST=localhost
export QUEUE_PORT=9092

# Postgres related setup
export DATABASE_HOST="localhost"
export DATABASE_PORT="5432"
export DATABASE_USER="postgres"
export DATABASE_PASSWORD="dev"
export DATABASE_NAME="sources_api_go_development"

# Redis related setup
export REDIS_CACHE_HOST="localhost"
export REDIS_CACHE_PORT="6379"

# Sources related setup
export BYPASS_RBAC=true
export PATH_PREFIX=/api
export APP_NAME="sources"
export PORT=4000
export METRICS_PORT=0

# Cloudigrade related setup
export KAFKA_SERVER_HOST=${QUEUE_HOST}
export KAFKA_SERVER_HOST=${QUEUE_PORT}
export SOURCES_API_BASE_URL="http://localhost:${PORT}"