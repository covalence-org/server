package requests

import (
	"errors"
	"netrunner/model"
	"netrunner/types"
	"time"

	"github.com/gin-gonic/gin"
)

// ModelInfo stores information about a registered model
type RegisterRequest struct {
	Name   string `json:"name" binding:"required"`
	Model  string `json:"model" binding:"required"`
	ApiUrl string `json:"api_url" binding:"required"`
}

func ParseRegisterRequest(raw *gin.Context) (model.Model, error) {

	var payload RegisterRequest
	if err := raw.ShouldBindJSON(&payload); err != nil {
		return model.Model{}, errors.New("invalid request format")
	}

	name, err := types.NewName(payload.Name)
	if err != nil {
		return model.Model{}, errors.New("invalid name")
	}

	modelName, err := types.NewModel(payload.Model)
	if err != nil {
		return model.Model{}, errors.New("invalid model")
	}

	apiUrl, err := types.NewApiUrl(payload.ApiUrl)
	if err != nil {
		return model.Model{}, errors.New("invalid api url")
	}

	return model.Model{
		Name:      name,
		Model:     modelName,
		ApiUrl:    apiUrl,
		CreatedAt: time.Now(),
	}, nil

}
