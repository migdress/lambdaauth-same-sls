package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/lambda-custom-authorizers/lambdaauth/internal/apigateway"
	"github.com/lambda-custom-authorizers/lambdaauth/internal/app/model"
	"github.com/lambda-custom-authorizers/lambdaauth/internal/app/repository/user"
	"github.com/lambda-custom-authorizers/lambdaauth/internal/passwordhasher"
	"github.com/pkg/errors"
)

type requestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Handler struct {
	userRepository userRepository
	passwordHasher passwordHasher
}

func NewHandler(
	userRepository userRepository,
	passwordHasher passwordHasher,
) *Handler {
	return &Handler{
		userRepository,
		passwordHasher,
	}
}

type userRepository interface {
	FindByEmail(email string) (*model.User, error)
	Save(user model.User) error
}

type passwordHasher interface {
	HashPassword(password string) (string, error)
}

func (h *Handler) Handle() func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		req := requestBody{}
		err := json.Unmarshal([]byte(request.Body), &req)
		if err != nil {
			return apigateway.Error(http.StatusBadRequest, err), nil
		}

		if req.Email == "" || req.Password == "" {
			return apigateway.Error(http.StatusBadRequest, errors.New("missing required fields")), nil
		}

		existing, err := h.userRepository.FindByEmail(req.Email)
		if err != nil && errors.Cause(err) != user.ErrUserNotFound {
			return apigateway.Error(http.StatusInternalServerError, err), nil
		}
		if existing != nil {
			return apigateway.Error(http.StatusBadRequest, errors.New("user already exists")), nil
		}

		hashedPassword, err := h.passwordHasher.HashPassword(req.Password)
		if err != nil {
			return apigateway.Error(http.StatusInternalServerError, err), nil
		}

		err = h.userRepository.Save(model.User{
			Email:    req.Email,
			Password: hashedPassword,
		})
		if err != nil {
			return apigateway.Error(http.StatusInternalServerError, err), nil
		}

		return apigateway.Respond(http.StatusOK, ""), nil
	}
}

func main() {
	tableName := os.Getenv("DYNAMODB_USER")
	if tableName == "" {
		panic("DYNAMODB_USER cannot be empty")
	}

	awsSession, err := session.NewSession()
	if err != nil {
		panic(err)
	}

	dynamodbClient := dynamodb.New(awsSession)

	userRepository := user.NewUserRepository(
		tableName,
		dynamodbClient,
	)

	passwordHasher := passwordhasher.NewPasswordHasher()

	handler := NewHandler(
		userRepository,
		passwordHasher,
	)

	lambda.Start(handler.Handle())
}
