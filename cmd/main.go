package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Blue-Onion/ArtmeisterBackend/config"
	"github.com/Blue-Onion/ArtmeisterBackend/handler"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/admin"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/art"
	artmetadata "github.com/Blue-Onion/ArtmeisterBackend/handler/artMetaData"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/event"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/logger"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/user"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

func main() {
	//Load Env
	cfg := config.GetConfig()

	//DB
	apiCfg, err := config.DbQuries()
	if err != nil {
		log.Fatalf("Couldn't connect to database: %v", err)
	}
	// Load Logger
	_, err = logger.GetLogger()
	if err != nil {
		log.Fatalf("Couldn't intialize Logger: %v", err)
	}
	//Handlers
	userHandler := &user.Handler{
		Repo: apiCfg.UserRepo,
	}
	middlewareHandler := &middleware.Handler{
		Repo: apiCfg.UserRepo,
	}
	artHanlder := &art.Handler{
		Repo: apiCfg.ArtRepo,
	}
	artMetaDataHandler := &artmetadata.Handler{
		Repo: apiCfg.ArtMetaDataRepo,
	}
	eventHandler := &event.EventHandler{
		Repo: apiCfg.EventRepo,
	}
	eventAttendeeHandler := &event.EventAttendeeHandler{
		Repo: apiCfg.EventAttendeeRepo,
	}

	//Server
	router := chi.NewRouter()
	AllowedOrigin := []string{
		fmt.Sprintf("http://%s", cfg.Frontend_Url),
		fmt.Sprintf("https://%s", cfg.Frontend_Url),
	}
	fmt.Println(AllowedOrigin)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   AllowedOrigin,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	router.Use(middleware.MiddlewareLogger)
	router.Use(middleware.MiddlewareRateLimit)
	router.Get("/health", handler.Health)
	router.Get("/", handler.MainPage)

	// User Routes
	userRoute := user.UserRouter(userHandler, middlewareHandler)
	router.Mount("/auth", userRoute)
	// Art Routes
	profile := art.ProfileHandler{
		UserRepo: apiCfg.UserRepo,
		ArtRepo:  apiCfg.ArtRepo,
	}
	artRoute := art.ArtRouter(artHanlder, artMetaDataHandler, middlewareHandler, &profile)
	router.Mount("/art", artRoute)
	// Event Routes
	eventRoute := event.EventRouter(eventHandler, eventAttendeeHandler, middlewareHandler)
	router.Mount("/event", eventRoute)
	// Admin Routes
	adminRoute := admin.AdminRoute(userHandler, artHanlder, middlewareHandler)
	router.Mount("/admin", adminRoute)

	server := http.Server{
		Handler: router,
		Addr:    "0.0.0.0:" + cfg.Port,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Listening on http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error occurred: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error occurred in Shutdown: %v", err)
	}
	log.Println("Server Shutdown gracefully")
}
