package dynamicdefense

import (
	"log"
	"net/http"

	promptrewriting "covalence/src/dynamic_defense/prompt_rewriting"
	"covalence/src/request"
	"covalence/src/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Defense struct {
	Enabled bool
	ID      uuid.UUID
	Type    types.DefenseType
}

func HookDefenses(c *gin.Context, payload *request.Generate) (int, error) {
	log.Printf("defense hook called with payload")
	// db := c.MustGet("db").(*postgres.DB)
	// requestID := c.MustGet("requestID").(string)

	// Check latest message
	err := promptrewriting.Apply(&payload.Messages)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
