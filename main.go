package main

import (
	"traineesheep/feedservice/models"
	"traineesheep/feedservice/utils"
	"traineesheep/feedservice/endpointHandlers"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load("development.env")

	models.InitDB()

	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // ограничение - 10 мб (3 картинки по 2 мб)
	})

	app.Static("/", "./static")

	// загрузка всей отсортированной ленты
	app.Get("/loadMainFeed", endpointHandlers.LoadMainFeedHandler)

	// загрузка отсортированной ленты пользователя
	app.Get("/loadUserFeed/:userID", endpointHandlers.LoadUserFeedHandler)

	// вход
	app.Post("/signin", endpointHandlers.SigninHandler)

	// регистрация
	app.Post("/signup", endpointHandlers.SignupHandler)

	// обновление токена
	app.Post("/refresh", endpointHandlers.RefreshHandler)

	// удаление токена у пользователя
	app.Post("/logout", endpointHandlers.LogoutHandler)

	// загрузка изображений
	app.Post("/upload", endpointHandlers.UploadHandler)
	
	app.Listen(utils.GetEnv("APP_HOST", ":3000"))
}
