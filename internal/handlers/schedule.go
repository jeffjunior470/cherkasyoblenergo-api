package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"cherkasyoblenergo-api/internal/config"
	"cherkasyoblenergo-api/internal/models"
	"cherkasyoblenergo-api/internal/utils"
)

func GetAPIInfo() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "cherkasyoblenergo-api",
			"version": strings.TrimPrefix(config.AppVersion, "v"),
			"endpoints": fiber.Map{
				"blackout-schedule": fiber.Map{
					"path":        "/cherkasyoblenergo/api/blackout-schedule",
					"method":      "GET",
					"description": "Get electricity blackout schedules",
					"parameters": []fiber.Map{
						{
							"name":        "option",
							"type":        "string",
							"required":    true,
							"description": "Filter option",
							"values":      []string{"all", "latest_n", "by_date", "by_schedule_date"},
						},
						{
							"name":        "date",
							"type":        "string",
							"required":    false,
							"description": "Date in YYYY-MM-DD format, or 'today', 'tomorrow'",
						},
						{
							"name":        "limit",
							"type":        "integer",
							"required":    false,
							"description": "Number of records to return (for latest_n and by_schedule_date options)",
						},
						{
							"name":        "formatted",
							"type":        "integer",
							"required":    false,
							"description": "Set to 1 for pretty-printed JSON with two-space indentation; otherwise JSON is compact",
							"values":      []int{0, 1},
						},
						{
							"name":        "queue",
							"type":        "string",
							"required":    false,
							"description": "Filter by specific queues (e.g. '1_1,2_1,3_2'). Format: X_Y where X is 1-6 and Y is 1-2",
						},
					},
					"examples": []fiber.Map{
						{
							"description": "Get all schedules",
							"url":         "/cherkasyoblenergo/api/blackout-schedule?option=all",
						},
						{
							"description": "Get latest 5 schedules",
							"url":         "/cherkasyoblenergo/api/blackout-schedule?option=latest_n&limit=5",
						},
						{
							"description": "Get schedule for today",
							"url":         "/cherkasyoblenergo/api/blackout-schedule?option=by_schedule_date&date=today",
						},
						{
							"description": "Get schedule for specific date with queue filter",
							"url":         "/cherkasyoblenergo/api/blackout-schedule?option=by_schedule_date&date=2025-12-25&queue=1_1,2_1",
						},
					},
				},
			},
			"source": "https://github.com/sigmanor/cherkasyoblenergo-api",
		})
	}
}

type ScheduleFilter struct {
	Option    string `json:"option"`
	Date      string `json:"date"`
	Limit     int    `json:"limit"`
	Queue     string `json:"queue"`
	Formatted bool   `json:"formatted"`
	Format    string `json:"format"`
}

func parseAndValidateQueues(queueStr string) ([]string, error) {
	if strings.TrimSpace(queueStr) == "" {
		return []string{}, nil
	}

	tokens := strings.Split(queueStr, ",")

	queuePattern := regexp.MustCompile(`^[1-6]_[1-2]$`)

	seen := make(map[string]bool)
	result := []string{}

	for _, token := range tokens {
		queue := strings.TrimSpace(token)

		if !queuePattern.MatchString(queue) {
			return nil, fmt.Errorf("invalid queue value: '%s'. Each queue must match format X_Y where X is 1-6 and Y is 1-2", queue)
		}

		if !seen[queue] {
			seen[queue] = true
			result = append(result, queue)
		}
	}

	return result, nil
}

func getQueueValue(schedule *models.Schedule, queueName string) string {
	switch queueName {
	case "1_1":
		return schedule.Col1_1
	case "1_2":
		return schedule.Col1_2
	case "2_1":
		return schedule.Col2_1
	case "2_2":
		return schedule.Col2_2
	case "3_1":
		return schedule.Col3_1
	case "3_2":
		return schedule.Col3_2
	case "4_1":
		return schedule.Col4_1
	case "4_2":
		return schedule.Col4_2
	case "5_1":
		return schedule.Col5_1
	case "5_2":
		return schedule.Col5_2
	case "6_1":
		return schedule.Col6_1
	case "6_2":
		return schedule.Col6_2
	default:
		return ""
	}
}

func buildFilteredResponse(schedules []models.Schedule, queueNames []string) []map[string]interface{} {
	result := make([]map[string]interface{}, len(schedules))
	for i, schedule := range schedules {
		resultMap := map[string]interface{}{
			"id":            schedule.ID,
			"news_id":       schedule.NewsID,
			"title":         schedule.Title,
			"date":          schedule.Date,
			"schedule_date": schedule.ScheduleDate,
		}

		for _, queueName := range queueNames {
			resultMap[queueName] = getQueueValue(&schedule, queueName)
		}

		result[i] = resultMap
	}
	return result
}

func resolveDateValue(dateStr string) string {
	now := time.Now()
	switch strings.ToLower(dateStr) {
	case "today":
		return now.Format("2006-01-02")
	case "tomorrow":
		return now.AddDate(0, 0, 1).Format("2006-01-02")
	default:
		return dateStr
	}
}

func handleScheduleRequest(c *fiber.Ctx, db *gorm.DB, filter ScheduleFilter) error {
	if filter.Format == "" {
		filter.Format = "json"
	}
	if filter.Format != "json" && filter.Format != "html" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid format parameter, expected 'json' or 'html'"})
	}

	filter.Date = resolveDateValue(filter.Date)

	var schedules []models.Schedule
	query := db.Table("schedules")
	switch filter.Option {
	case "all":
		if err := query.Find(&schedules).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve records"})
		}
	case "latest_n":
		if filter.Limit <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid limit value, it must be greater than zero"})
		}
		if err := query.Order("date desc, news_id desc").Limit(filter.Limit).Find(&schedules).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve latest records"})
		}
	case "by_date":
		if filter.Date == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Date must be specified in YYYY-MM-DD format"})
		}
		date, err := time.Parse("2006-01-02", filter.Date)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format, expected YYYY-MM-DD"})
		}
		if err := query.Where("DATE(date) = ?", date.Format("2006-01-02")).Find(&schedules).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve records by date"})
		}
	case "by_schedule_date":
		if filter.Date == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Date must be specified in YYYY-MM-DD format"})
		}
		if _, err := time.Parse("2006-01-02", filter.Date); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format, expected YYYY-MM-DD"})
		}
		if filter.Limit < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid limit value, it must be greater than or equal to zero"})
		}

		scheduleQuery := db.Table("schedules").Where("schedule_date = ?", filter.Date).Order("date desc, news_id desc")
		if filter.Limit > 0 {
			scheduleQuery = scheduleQuery.Limit(filter.Limit)
		}
		if err := scheduleQuery.Find(&schedules).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve records"})
		}

		neededMore := filter.Limit == 0 || len(schedules) < filter.Limit
		if neededMore {
			var legacySchedules []models.Schedule
			legacyQuery := db.Table("schedules").Where("schedule_date = '' OR schedule_date IS NULL").Order("date desc, news_id desc")
			if err := legacyQuery.Find(&legacySchedules).Error; err == nil {
				for i := range legacySchedules {
					legacySchedules[i].ScheduleDate = utils.ExtractScheduleDateFromTitle(legacySchedules[i].Title)
					if legacySchedules[i].ScheduleDate == filter.Date {
						schedules = append(schedules, legacySchedules[i])
						if filter.Limit > 0 && len(schedules) >= filter.Limit {
							break
						}
					}
				}
			}
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid option parameter value"})
	}

	for i := range schedules {
		if schedules[i].ScheduleDate == "" {
			schedules[i].ScheduleDate = utils.ExtractScheduleDateFromTitle(schedules[i].Title)
		}
	}

	validatedQueues, err := parseAndValidateQueues(filter.Queue)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if filter.Format == "html" {
		queues := validatedQueues
		if len(queues) == 0 {
			queues = []string{"1_1", "1_2", "2_1", "2_2", "3_1", "3_2", "4_1", "4_2", "5_1", "5_2", "6_1", "6_2"}
		}
		page := buildScheduleHTML(schedules, queues)
		c.Type("html")
		return c.SendString(page)
	}

	var response interface{}
	if len(validatedQueues) == 0 {
		response = schedules
	} else {
		response = buildFilteredResponse(schedules, validatedQueues)
	}

	if filter.Formatted {
		body, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return err
		}
		c.Type("json")
		return c.Send(body)
	}
	return c.JSON(response)
}

func buildScheduleHTML(schedules []models.Schedule, queues []string) string {
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html lang="uk"><head><meta charset="utf-8"><title>Графіки відключень</title><link rel="stylesheet" href="/static/schedule.css"></head><body>`)
	b.WriteString(`<h1>Графіки відключень електроенергії</h1>`)
	b.WriteString(`<p><a href="https://www.cherkasyoblenergo.com/">Новини Черкасиобленерго</a></p>`)
	b.WriteString(`<table border="1" cellpadding="4" cellspacing="0"><thead><tr><th>Дата</th><th>ID новини</th><th>Розклад на</th><th>Заголовок</th></tr></thead><tbody>`)

	for i := range schedules {
		sch := &schedules[i]
		if sch.ScheduleDate == "" {
			sch.ScheduleDate = utils.ExtractScheduleDateFromTitle(sch.Title)
		}
		b.WriteString(`<tr><td>`)
		b.WriteString(html.EscapeString(formatDateOrDateTime(sch.Date)))
		b.WriteString(`</td><td>`)
		b.WriteString(html.EscapeString(fmt.Sprintf("%d", sch.NewsID)))
		b.WriteString(`</td><td>`)
		b.WriteString(html.EscapeString(formatScheduleDate(sch.ScheduleDate)))
		b.WriteString(`</td><td>`)
		b.WriteString(html.EscapeString(sch.Title))
		b.WriteString(`</td></tr>`)
		b.WriteString(`<tr class="nested-header"><td></td><td colspan="2">Номер черги</td><td>Графік</td></tr>`)
		for _, q := range queues {
			queueLabel := "Черга " + strings.ReplaceAll(q, "_", ".")
			b.WriteString(`<tr class="nested-row"><td></td><td colspan="2">`)
			b.WriteString(html.EscapeString(queueLabel))
			b.WriteString(`</td><td>`)
			b.WriteString(html.EscapeString(getQueueValue(sch, q)))
			b.WriteString(`</td></tr>`)
		}
	}
	b.WriteString(`</tbody></table></body></html>`)
	return b.String()
}

func formatDateOrDateTime(t time.Time) string {
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return t.Format("02.01.2006")
	}
	return t.Format("02.01.2006 15:04")
}

func formatScheduleDate(value string) string {
	if value == "" {
		return ""
	}
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return date.Format("02.01.2006")
}

func GetSchedule(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter := ScheduleFilter{
			Option:    c.Query("option"),
			Date:      c.Query("date"),
			Queue:     c.Query("queue"),
			Formatted: c.Query("formatted") == "1",
			Format:    c.Query("format", "json"),
		}

		if limitStr := c.Query("limit"); limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid limit parameter, expected integer value"})
			}
			filter.Limit = limit
		}

		return handleScheduleRequest(c, db, filter)
	}
}
