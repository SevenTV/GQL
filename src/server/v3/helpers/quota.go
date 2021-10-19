package helpers

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
)

type Quota struct {
	C *fiber.Ctx
	// The initial amount of Quota Points
	Limit *int32 `json:"limit"`
	// Quota Points available
	Points *int32 `json:"points"`
	// Map of fields to quota costs
	Fields sync.Map
}

func (q *Quota) DecreaseByOne(object string, field string) bool {
	return q.Decrease(1, object, field)
}

func (q *Quota) Decrease(amount int32, object string, field string) bool {
	atomic.AddInt32(q.Points, -amount)

	fieldName := fmt.Sprintf("%s.%s", object, field)
	current, ok := q.Fields.Load(fieldName)
	if !ok {
		q.Fields.Store(fieldName, amount)
	} else {
		q.Fields.Store(fieldName, current.(int32)+amount)
	}

	return q.Check()
}

func (q *Quota) Check() bool {
	return q.GetPoints() >= 0
}

func (q *Quota) GetPoints() int32 {
	return atomic.LoadInt32(q.Points)
}

func (q *Quota) GetLimit() int32 {
	return atomic.LoadInt32(q.Limit)
}
