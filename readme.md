Пуск: 
docker build -t my-go-rest .
docker run -d -p 8080:8080 --name my-rest my-go-rest