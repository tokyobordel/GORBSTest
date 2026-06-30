Пуск:
docker network create chatnet

docker run -d --name postgres --network chatnet -e POSTGRES_PASSWORD=123 postgres:16-alpine

docker run -d --name chatapp --network chatnet -p 8080:8080 -e DB_HOST=postgres -e DB_PASSWORD=123 myrest