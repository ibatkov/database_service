#!/bin/bash
file_paths=("./docker-compose.yml.example" "./config/database-service/config.yml.example" "./config/migration/goose.env.example" "./config/prometheus/prometheus.yml.example" "./config/postgres/config.env.example" "./config/redis/redis.conf.example")
for file_path in "${file_paths[@]}"
do
  if [ -f "$file_path" ]; then
    new_file_path="${file_path%.example}"
    cp "$file_path" "$new_file_path"
    echo "Файл $file_path переименован в $new_file_path"
  else
    echo "Файл $file_path не найден"
  fi
done