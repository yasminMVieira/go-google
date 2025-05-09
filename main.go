package main

import (
	"go-google/config"
	"go-google/handlers"
	"go-google/middleware"
	"go-google/models"
	"go-google/repository"
	"go-google/services"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Carregar configurações
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Configurar banco de dados
	db, err := config.SetupDatabase(cfg)
	if err != nil {
		log.Fatalf("Erro ao configurar banco de dados: %v", err)
	}

	// Auto-migrar modelos
	err = db.AutoMigrate(&models.User{}, &models.Group{}, &models.Role{})
	if err != nil {
		log.Fatalf("Erro ao migrar banco de dados: %v", err)
	}

	// Inicializar repositórios
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)

	// Inicializar serviços
	authService := services.NewAuthService(cfg, userRepo, groupRepo)
	userService := services.NewUserService(userRepo, groupRepo)

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)

	// Configurar router
	router := gin.Default()

	// Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Rotas de autenticação (públicas)
	auth := router.Group("/auth")
	{
		auth.GET("/login", authHandler.GoogleLogin)
		auth.GET("/callback", authHandler.GoogleCallback)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Rotas protegidas (requerem autenticação)
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Rotas de usuário
		api.GET("/profile", userHandler.GetProfile)
		
		// Rotas administrativas (requerem role específica)
		admin := api.Group("/admin")
		admin.Use(middleware.RoleMiddleware("admin"))
		{
			admin.GET("/users", userHandler.ListUsers)
			admin.POST("/groups", userHandler.CreateGroup)
			admin.PUT("/users/:id/groups", userHandler.AssignUserToGroup)
		}
	}

	// Iniciar servidor
	log.Printf("Servidor iniciado na porta %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}