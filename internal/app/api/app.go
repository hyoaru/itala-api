package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	handler "github.com/hyoaru/itala-api/internal/app/api/handler"
	category "github.com/hyoaru/itala-api/internal/features/category"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type App struct{ server *http.Server }

func New(addr string) *App {
	dynamodbClient := dynamodbclient.NewSDKDynamoDBClient()

	identityProvider := identity.NewCognitoIdentityProvider(os.Getenv("AWS_REGION"), os.Getenv("COGNITO_USER_POOL_ID"))
	categoryRepository := category.NewDynamoDBCategoryRepository(dynamodbClient, os.Getenv("DYNAMODB_TABLE_NAME"))

	categoryHandler := &handler.CategoryHandler{CreateCategory: category.NewCreateCategory(categoryRepository)}
	router := NewRouter(identityProvider, *categoryHandler)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	return &App{server: server}
}

func (app *App) Run() error {
	logger.Debug(fmt.Sprintf("Server has started at %s", app.server.Addr))
	return app.server.ListenAndServe()
}
