package endpointHandlers

import (
	"traineesheep/feedservice/models"
	"traineesheep/feedservice/utils"
	"github.com/gofiber/fiber/v2"
	"time"
	"strings"
	"fmt"
	"net/http"
	"bytes"
	"mime/multipart"
	"io"
	"log"
	"encoding/json"
)

func UploadHandler(c *fiber.Ctx) error {
	// Проверяем Content-Type
    contentType := string(c.Request().Header.ContentType())
    if !strings.HasPrefix(contentType, "multipart/form-data") {
        return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
            Success: false, ErrMessage: "Ожидается multipart/form-data",
        })
    }

    // Парсим форму
    form, err := c.MultipartForm()
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
            Success: false, ErrMessage: "Неверный формат данных",
        })
    }

    files := form.File["images"] // поле, в котором фронт отправляет файлы
    if len(files) == 0 {
        return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
            Success: false, ErrMessage: "Не выбрано ни одного изображения",
        })
    }
    if len(files) > 3 {
        return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
            Success: false, ErrMessage: "Максимальное количество изображений — 3",
        })
    }

    const maxFileSize = 2 << 20 // 2 МБ
    for _, file := range files {
        if file.Size > maxFileSize {
            return c.Status(fiber.StatusBadRequest).JSON(utils.ApiResponse{
                Success: false,
                ErrMessage: fmt.Sprintf("Файл '%s' превышает 2 МБ", file.Filename),
            })
        }
    }

    // URL внешнего сервиса
    imageAddURL := utils.GetEnv("IMAGE_ADD_URL", "")
    if imageAddURL == "" {
        return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
            Success: false, ErrMessage: "ImageService не настроен",
        })
    }

    // Структура ответа от внешнего сервиса (предполагаем {"id": 123})
    type externalImageResponse struct {
        ID int `json:"id"`
    }

    var imageIDs []int // сюда соберём полученные id

    // Отправляем каждый файл во внешний сервис
    for _, fileHeader := range files {
        file, err := fileHeader.Open()
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
                Success: false, ErrMessage: "Не удалось прочитать файл",
            })
        }
        defer file.Close()

        // Готовим multipart-запрос
        var b bytes.Buffer
        writer := multipart.NewWriter(&b)
        part, err := writer.CreateFormFile("file", fileHeader.Filename) // поле "file"
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
                Success: false, ErrMessage: "Ошибка подготовки запроса к ImageService",
            })
        }
        _, err = io.Copy(part, file)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
                Success: false, ErrMessage: "Ошибка чтения файла",
            })
        }
        writer.Close()

        // HTTP POST на внешний сервис
        httpReq, err := http.NewRequest("POST", imageAddURL, &b)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
                Success: false, ErrMessage: "Внутренняя ошибка",
            })
        }
        httpReq.Header.Set("Content-Type", writer.FormDataContentType())

        client := &http.Client{Timeout: 10 * time.Second}
        resp, err := client.Do(httpReq)
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
                Success: false, ErrMessage: "Не удалось сохранить изображение в ImageService",
            })
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
            bodyBytes, _ := io.ReadAll(resp.Body)
            log.Printf("Ошибка внешнего сервиса: статус %d, тело %s", resp.StatusCode, string(bodyBytes))
            return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
                Success: false, ErrMessage: "ImageService вернул ошибку",
            })
        }

        var extResp externalImageResponse
        if err := json.NewDecoder(resp.Body).Decode(&extResp); err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
                Success: false, ErrMessage: "Неверный ответ от ImageService",
            })
        }

        imageIDs = append(imageIDs, extResp.ID)
    }

    // Все файлы успешно отправлены, создаём пост и записи в image_post
    tx, err := models.DB.Begin()
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
            Success: false, ErrMessage: "Ошибка базы данных",
        })
    }
    defer tx.Rollback()

    var post models.Post
    err = tx.QueryRow(
        "INSERT INTO post (user_id, title, description) VALUES ($1, $2, $3) RETURNING id, user_id, title, description, created_at",
        0, "Загрузка изображений", "", // user_id = 0, пока нет авторизации
    ).Scan(&post.ID, &post.UserID, &post.Title, &post.Description, &post.CreatedAt)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
            Success: false, ErrMessage: "Не удалось создать пост",
        })
    }

    for _, imgID := range imageIDs {
        _, err = tx.Exec(
            "INSERT INTO image_post (post_id, image_id) VALUES ($1, $2)",
            post.ID, imgID,
        )
        if err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
                Success: false, ErrMessage: "Ошибка привязки изображения к посту",
            })
        }
    }

    if err := tx.Commit(); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(utils.ApiResponse{
            Success: false, ErrMessage: "Ошибка сохранения данных",
        })
    }

	// todo отправить уведомление

    return c.Status(fiber.StatusCreated).JSON(utils.ApiResponse{
        Success: true,
        Data: fiber.Map{
            "post":      post,
            "image_ids": imageIDs,
        },
    })
}