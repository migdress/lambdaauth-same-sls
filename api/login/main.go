package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/migdress/lambdaauthsamesls/internal/apigateway"
	"github.com/migdress/lambdaauthsamesls/internal/app/model"
	"github.com/migdress/lambdaauthsamesls/internal/app/repository/user"
	"github.com/migdress/lambdaauthsamesls/internal/jwtwrapper"
	"github.com/migdress/lambdaauthsamesls/internal/passwordhasher"
	"github.com/pkg/errors"
)

type requestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type responseBody struct {
	JWT string `json:"jwt"`
}

type Handler struct {
	userRepository userRepository
	passwordHasher passwordHasher
	jwtWrapper     jwtWrapper
}

func NewHandler(
	userRepository userRepository,
	passwordHasher passwordHasher,
	jwtWrapper jwtWrapper,
) *Handler {
	return &Handler{
		userRepository,
		passwordHasher,
		jwtWrapper,
	}
}

type userRepository interface {
	FindByEmail(email string) (*model.User, error)
	Save(user model.User) error
}

type passwordHasher interface {
	VerifyPassword(password string, hashedPassword string) error
}

type jwtWrapper interface {
	GenerateJWT(customClaims map[string]string) (string, error)
}

func (h *Handler) Handle() func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		req := requestBody{}
		err := json.Unmarshal([]byte(request.Body), &req)
		if err != nil {
			return apigateway.Error(http.StatusBadRequest, err), nil
		}

		if req.Email == "" || req.Password == "" {
			return apigateway.Error(http.StatusUnauthorized, errors.New("Unauthorized")), nil
		}

		existing, err := h.userRepository.FindByEmail(req.Email)
		if err != nil {
			return apigateway.Error(http.StatusUnauthorized, errors.Wrap(err, "Unauthorized")), nil
		}

		err = h.passwordHasher.VerifyPassword(req.Password, existing.Password)
		if err != nil {
			return apigateway.Error(http.StatusUnauthorized, errors.Wrap(err, "Unauthorized")), nil
		}

		token, err := h.jwtWrapper.GenerateJWT(map[string]string{
			"email": existing.Email,
		})

		response := responseBody{
			JWT: token,
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			return apigateway.Error(http.StatusInternalServerError, err), nil
		}

		return apigateway.Respond(http.StatusOK, string(responseBytes)), nil
	}
}

func main() {
	tableName := os.Getenv("DYNAMODB_USER")
	if tableName == "" {
		panic("DYNAMODB_USER cannot be empty")
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		panic("SECRET_KEY cannot be empty")
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

	jwtWrapper := jwtwrapper.NewJWTWrapper(
		secretKey,
		time.Minute*15,
	)

	handler := NewHandler(
		userRepository,
		passwordHasher,
		jwtWrapper,
	)

	lambda.Start(handler.Handle())
}
