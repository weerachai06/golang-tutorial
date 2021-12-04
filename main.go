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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/weerachai06/todo/auth"
	"github.com/weerachai06/todo/todo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/* type User struct {
	gorm.Model
	Name string
}
*/

var (
	buildcommit = "dev"
	buildtime   = time.Now().String()
)

func main() {
	err := godotenv.Load("local.env")
	if err != nil {
		log.Printf("Please consider environment variables: %s", err)
	}
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to conntect database")
	}

	db.AutoMigrate(&todo.Todo{})

	//db.Create(&User{Name: "Petch"})
	r := gin.Default()
	//userHandler := UserHandler{db: db}

	//r.GET("users", userHandler.User)
	r.GET("/x", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"buildcommit": buildcommit,
			"buildtime":   buildtime,
		})
	})
	r.GET("/tokenz", auth.AccessToken(os.Getenv("SIGN")))
	protected := r.Group("", auth.Protect([]byte(os.Getenv("SIGN"))))
	handler := todo.NewTodoHandler(db)

	protected.POST("/todos", handler.NewTask)

	//r.Run()

	//Graceful Shuting Down
	s := &http.Server{
		Addr:           fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	stop()

	fmt.Println("shutting down gracefully, press Ctrl+C again to force")
	timeOutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(timeOutCtx); err != nil {
		fmt.Println(err)
	}
}

/* type UserHandler struct {
	db *gorm.DB
} */

/* func (h *UserHandler) User(c *gin.Context) {
	var user User
	h.db.First(&user)
	c.JSON(200, user)
}
*/
