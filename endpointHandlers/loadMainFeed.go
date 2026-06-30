package endpointHandlers

import (
	"traineesheep/feedservice/models"
	"traineesheep/feedservice/utils"
	"github.com/gofiber/fiber/v2"
)

func LoadMainFeedHandler(c *fiber.Ctx) error {
	rows, err := models.DB.Query(`
        SELECT p.id, p.user_id, COALESCE(u.username, '') as username,
               p.title, p.description, 
			   TO_CHAR(p.created_at, 'DD.MM.YYYY HH24:MI:SS') as created_at
        FROM post p
        LEFT JOIN users u ON p.user_id = u.id
        ORDER BY p.created_at DESC
    `)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
            Success: false, ErrMessage: "Ошибка получения постов",
        })
    }
    defer rows.Close()

    var posts []models.Post
    for rows.Next() {
        var p models.Post
        if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, 
			&p.Description, &p.CreatedAt); err != nil {
            continue
        }
        posts = append(posts, p)
    }

    return c.JSON(utils.ApiResponse{
        Success: true,
        Data:    posts,
    })
}