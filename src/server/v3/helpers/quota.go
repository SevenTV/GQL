package helpers

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
)

type Quota struct {
	lock sync.Mutex
	C    *fiber.Ctx
	// The initial amount of Quota Points
	Limit int `json:"limit"`
	// Quota Points available
	Points int `json:"points"`
	// Map of fields to quota costs
	Fields map[string]int
}

func (q *Quota) DecreaseByOne(object string, field string) bool {
	return q.Decrease(1, object, field)
}

func (q *Quota) Decrease(amount int, object string, field string) bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.Points -= amount

	fieldName := fmt.Sprintf("%s.%s", object, field)
	q.Fields[fieldName] += amount

	return q.Check()
}

func (q *Quota) Check() bool {
	return q.Points >= 0
}
