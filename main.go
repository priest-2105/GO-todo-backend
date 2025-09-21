package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"

    "github.com/joho/godotenv"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type Task struct {
    ID          int    `json:"id" gorm:"primaryKey"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Done        bool   `json:"done"`
}

func main() {
    _ = godotenv.Load()
    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_SSLMODE"),
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("failed to connect:", err)
    }

    if err := db.AutoMigrate(&Task{}); err != nil {
        log.Fatal("migration failed:", err)
    }

    // === List All ===
    http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
        var tasks []Task
        if result := db.Find(&tasks); result.Error != nil {
            http.Error(w, result.Error.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(tasks)
    })

    // === Add New ===
    http.HandleFunc("/todos/add", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var task Task
        if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        if result := db.Create(&task); result.Error != nil {
            http.Error(w, result.Error.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(task)
    })

    // === View One ===
    http.HandleFunc("/todos/view/", func(w http.ResponseWriter, r *http.Request) {
        id := strings.TrimPrefix(r.URL.Path, "/todos/view/")
        var task Task
        if result := db.First(&task, id); result.Error != nil {
            http.Error(w, "Task not found", http.StatusNotFound)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(task)
    })

    // === Delete ===
    http.HandleFunc("/todos/delete/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodDelete {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        id := strings.TrimPrefix(r.URL.Path, "/todos/delete/")
        if result := db.Delete(&Task{}, id); result.Error != nil {
            http.Error(w, result.Error.Error(), http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"message":"Deleted"}`))
    })

    // === Update Title/Description ===
    http.HandleFunc("/todos/update/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPut {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        id := strings.TrimPrefix(r.URL.Path, "/todos/update/")

        var payload Task
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        var task Task
        if result := db.First(&task, id); result.Error != nil {
            http.Error(w, "Task not found", http.StatusNotFound)
            return
        }

        task.Title = payload.Title
        task.Description = payload.Description
        if result := db.Save(&task); result.Error != nil {
            http.Error(w, result.Error.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(task)
    })

    // === Mark as Done ===
    http.HandleFunc("/todos/done/", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPatch {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        id := strings.TrimPrefix(r.URL.Path, "/todos/done/")

        var task Task
        if result := db.First(&task, id); result.Error != nil {
            http.Error(w, "Task not found", http.StatusNotFound)
            return
        }

        task.Done = true
        if result := db.Save(&task); result.Error != nil {
            http.Error(w, result.Error.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(task)
    })

    log.Println("Server running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
