package router

import (
	"covalence/src/register"
	"covalence/src/request"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterModel(c *gin.Context) {

	r := c.MustGet("registry").(*register.Registry)

	// Parse Request
	modelInfo, err := request.ParseRegister(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r.Register(modelInfo)
	log.Printf("Model registered: %s -> %s at %s", modelInfo.Name.String(), modelInfo.Model.String(), modelInfo.APIURL.String())

	c.JSON(http.StatusOK, gin.H{"status": "model registered", "name": modelInfo.Name.String(), "model": modelInfo.Model.String()})
}

func ListRegisteredModels(c *gin.Context) {
	r := c.MustGet("registry").(*register.Registry)

	r.Mu.RLock()
	defer r.Mu.RUnlock()

	models := make([]map[string]string, 0, len(r.Models))
	for _, info := range r.Models {
		models = append(models, map[string]string{
			"name":          info.Name.String(),
			"model":         info.Model.String(),
			"registered_at": info.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}
