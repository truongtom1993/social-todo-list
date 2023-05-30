package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
	"time"
)

type TodoItem struct {
	Id          int        `json:"id" gorm:"column:id;"`
	Title       string     `json:"title" gorm:"column:title;"`
	Description string     `json:"description" gorm:"column:description;"`
	Status      string     `json:"status" gorm:"column:status;"`
	CreatedAt   *time.Time `json:"created_at" gorm:"column:created_at;"`
	UpdatedAt   *time.Time `json:"updated_at" gorm:"column:updated_at;"`
}

type TodoItemCreation struct {
	Id          int    `json:"-" gorm:"column:id;"`
	Title       string `json:"title" gorm:"column:title;"`
	Description string `json:"description" gorm:"column:description;"`
}
type TodoItemUpdate struct {
	Title       string `json:"title" gorm:"column:title;"`
	Description string `json:"description" gorm:"column:description;"`
	Status      string `json:"status" gorm:"column:status;"`
}

func (TodoItem) TableName() string         { return "todo_items" }
func (TodoItemCreation) TableName() string { return TodoItem{}.TableName() }
func (TodoItemUpdate) TableName() string   { return TodoItem{}.TableName() }

func main() {
	dsn := os.Getenv("DB_CONN_STR")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Khong ket noi duoc voi DB")
	}

	r := gin.Default()
	//CURD
	//POST : /v1/items/ (create new item)
	//GET: /v1/items/ (get list all items) /v1/items?page=1
	//GET: /v1/items/:id (get detail item by id)
	//(PUT | PATCH): /v1/items/:id
	//DELETE: /v1/items/:id

	v1 := r.Group("/v1/")
	{
		items := v1.Group("items/")
		{
			items.POST("", CreateItem(db))
			items.GET("")
			items.GET(":id", GetItem(db))
			items.PATCH(":id", UpdateItem(db))
			items.DELETE(":id")
		}
	}

	if err := r.Run(":3131"); err != nil {
		panic("khong start duoc server")
	}
}

func CreateItem(db *gorm.DB) func(*gin.Context) {
	return func(context *gin.Context) {
		var data = new(TodoItemCreation)
		if err := context.ShouldBindJSON(data); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "🎁 main.go Line:70 ID:263de8, Dữ liệu truyền lên không hợp lệ",
			})
			return
		}
		if err := db.Create(data).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "🎁 main.go Line:76 ID:e03cb7, Không tạo được đối tượng",
			})
			return
		}
		context.JSON(http.StatusOK, gin.H{
			"data": data,
		})
	}
}
func GetItem(db *gorm.DB) func(*gin.Context) {
	return func(context *gin.Context) {
		var data = new(TodoItem)
		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "🎁 main.go Line:92 ID:b1484a, id không hợp lệ",
			})
			return
		}
		data.Id = id
		if err := db.First(data).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "🎁 main.go Line:100 ID:9b1935, Không tìm thấy dữ liệu",
			})
			return
		}
		context.JSON(http.StatusOK, gin.H{
			"data": data,
		})
	}
}
func UpdateItem(db *gorm.DB) func(*gin.Context) {
	return func(context *gin.Context) {
		var data = new(TodoItemUpdate)
		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "🎁 main.go Line:92 ID:b1484a, id không hợp lệ",
			})
			return
		}
		if err := context.ShouldBindJSON(data); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "🎁 main.go Line:70 ID:263de8, Dữ liệu truyền lên không hợp lệ",
			})
			return
		}
		if err := db.Where("id = ?", id).Updates(data).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "🎁 main.go Line:100 ID:9b1935, Cập nhật không thành công",
			})
			return
		}
		context.JSON(http.StatusOK, gin.H{
			"updated": true,
		})
	}
}
