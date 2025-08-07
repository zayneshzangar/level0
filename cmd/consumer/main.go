package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"order/config"
	v1 "order/internal/controller/http/v1"
	"order/internal/controller/kafka"
	"order/internal/service"
	"order/internal/storage"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка env: %v", err)
	}

	// Проверка Kafka конфигурации
	log.Printf("Kafka host: %s, ports: %s,%s,%s, topic: %s, group: %s",
		cfg.Kafka.Host, cfg.Kafka.Port1, cfg.Kafka.Port2, cfg.Kafka.Port3,
		cfg.Kafka.Topic, cfg.Kafka.GroupName)

	// Формирование bootstrap.servers
	var builder strings.Builder
	fmt.Fprintf(&builder, "%s:%s,%s:%s,%s:%s",
		cfg.Kafka.Host, cfg.Kafka.Port1,
		cfg.Kafka.Host, cfg.Kafka.Port2,
		cfg.Kafka.Host, cfg.Kafka.Port3,
	)
	bootstrapServers := builder.String()
	log.Printf("Kafka bootstrap.servers: %s", bootstrapServers)
	if bootstrapServers == "" {
		log.Fatal("Kafka bootstrap.servers is empty")
	}

	// Получаем *sql.DB и repo
	db, repo, err := storage.NewDatabaseConnection(cfg)
	if err != nil {
		log.Fatalf("Ошибка при подключении к базе данных: %v", err)
	}
	defer db.Close()

	// Инициализация базы
	if err := storage.InitDB(context.Background(), db); err != nil {
		log.Fatalf("Ошибка при инициализации базы: %v", err)
	}

	// Создание сервиса
	svc := service.NewService(repo)

	// Загрузка кэша из БД
	if err := svc.LoadCacheFromDB(context.Background()); err != nil {
		log.Fatalf("Failed to load cache from DB: %v", err)
	}

	// Создание Kafka-контроллера
	kafkaCtrl, err := kafka.NewKafkaController(
		bootstrapServers,
		cfg.Kafka.GroupName,
		cfg.Kafka.Topic,
		svc,
	)
	if err != nil {
		log.Fatalf("Failed to create Kafka controller: %v", err)
	}
	defer kafkaCtrl.Close()

	// Создание HTTP-хендлера и роутера
	handler := v1.NewHandler(svc)
	router := v1.NewRouter(handler, cfg)
	cors := v1.Cors(router, cfg)

	// Контекст для грациозного завершения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск консюмера в отдельной горутине
	go func() {
		if err := kafkaCtrl.Consume(ctx); err != nil {
			log.Printf("Kafka consumer stopped: %v", err)
		}
	}()

	// Запуск HTTP-сервера
	server := &http.Server{
		Addr:    cfg.App.Port,
		Handler: cors,
	}
	go func() {
		log.Println("Starting REST API server on", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидание сигналов для грациозного завершения
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// Отмена контекста для остановки консюмера
	cancel()

	// Остановка HTTP-сервера
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Application stopped")
}
