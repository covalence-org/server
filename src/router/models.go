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

	err = r.Register(modelInfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("model registered: %s -> %s at %s", modelInfo.Name.String(), modelInfo.Model.String(), modelInfo.APIURL.String())
	log.Println("model status set to: active")
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
			"status":        info.Status.String(),
			"provider":      info.Provider.String(),
			"api_url":       info.APIURL.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"models": models})
}

func ListModelProviders(c *gin.Context) {

	r := c.MustGet("providers").(*[]register.ModelProvider)

	modelProviders := make([]map[string]interface{}, 0, len(*r))
	for _, provider := range *r {

		// turn models into list of string
		models := make([]string, 0, len(provider.Models))
		for _, model := range provider.Models {
			models = append(models, model.String())
		}

		modelProviders = append(modelProviders, map[string]interface{}{
			"provider": provider.Provider.String(),
			"models":   models,
			"api_url":  provider.APIURL.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"providers": modelProviders})
}
