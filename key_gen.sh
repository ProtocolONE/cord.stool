#!/bin/bash

openssl genrsa -out ./service/config/keys/private_key 2048
openssl rsa -pubout -in ./service/config/keys/private_key -out ./service/config/keys/public_key.pub
