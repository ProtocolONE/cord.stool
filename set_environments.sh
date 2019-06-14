#!/bin/bash

export TRACKER_URL="http://127.0.0.1:5002"
export TRACKER_USER="admin"
export TRACKER_PASSWORD="123456"
export MONGO_HOST="127.0.0.1:27017"
export MONGO_DB="cord_stool"
export MONGO_USER=""
export MONGO_PASSWORD=""
export HTTP_SCHEME="http"
export SERVICE_PORT=5001
export PRIVATE_KEY_PATH="/home/testapp/test/cord.stool/service/config/keys/private_key"
export PUBLIC_KEY_PATH="/home//testapp/test/cord.stool/service/config/keys/public_key.pub"
export JWT_EXPIRATION_DELTA=30
export JWT_REFRESH_EXPIRATION_DELTA=72
export STORAGE_ROOT_PATH="/home/testapp/test/server_storage"

./cord.stool service
