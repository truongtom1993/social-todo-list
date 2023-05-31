package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	Title       *string `json:"title" gorm:"column:title;"`
	Description *string `json:"description" gorm:"column:description;"`
	Status      *string `json:"status" gorm:"column:status;"`
}

type Paging struct {
	Page  int   `json:"page" form:"page"`
	Limit int   `json:"limit" form:"limit"`
	Total int64 `json:"total" form:"-"`
}

func (paging *Paging) Process() {
	if paging.Page <= 0 {
		paging.Page = 1
	}
	if paging.Limit <= 0 || paging.Limit >= 100 {
		paging.Limit = 10
	}
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

	ginEngine := gin.Default()
	//CURD
	//POST : /v1/items/ (create new item)
	//GET: /v1/items/ (get list all items) /v1/items?page=1&limit=10
	//GET: /v1/items/:id (get detail item by id)
	//(PUT | PATCH): /v1/items/:id
	//DELETE: /v1/items/:id

	v1 := ginEngine.Group("/v1/")
	{
		items := v1.Group("items/")
		{
			items.POST("", CreateItem(db))
			items.GET("", ListItem(db))
			items.GET(":id", GetItem(db))
			items.PATCH(":id", UpdateItem(db))
			items.DELETE(":id", DeleteItem(db))
		}
	}

	if err := ginEngine.Run(":3131"); err != nil {
		panic("khong start duoc server")
	}
}

func CreateItem(db *gorm.DB) func(*gin.Context) {
	return func(context *gin.Context) {
		var data = new(TodoItemCreation)
		if err := context.ShouldBindJSON(data); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go Line:106 ID:94ffd9",
			})
			return
		}
		if err := db.Create(data).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go Line:113 ID:b7fa8e",
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
				"error":  err.Error(),
				"detail": "🎁 main.go Line:129 ID:a506e4",
			})
			return
		}
		data.Id = id
		if err := db.First(data).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go Line:137 ID:2e7d2a",
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
				"error":  err.Error(),
				"detail": "🎁 main.go Line:153 ID:ed799c",
			})
			return
		}
		if err := context.ShouldBindJSON(data); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go Line:160 ID:6b83d6",
			})
			return
		}
		if err := db.Where("id = ?", id).Updates(data).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go Line:100 ID:9b1935, Cập nhật không thành công",
			})
			return
		}
		context.JSON(http.StatusOK, gin.H{
			"updated": true,
		})
	}
}
func ListItem(db *gorm.DB) func(context *gin.Context) {
	return func(context *gin.Context) {
		//Parse tu param
		paging := new(Paging)
		if err := context.ShouldBind(paging); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go Line:182 ID:6e8c0f, Dữ liệu truyền lên không đúng định dạng",
			})
		}

		paging.Process()

		if err := db.Raw(fmt.Sprintf("select count(*) from %v", TodoItem{}.TableName())).Scan(&paging.Total).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "Khong dem duoc",
			})
		}

		data := new([]TodoItem)
		pageLimit := paging.Limit
		pageOffset := (paging.Page - 1) * pageLimit
		if err := db.Raw(fmt.Sprintf("SELECT * FROM %v WHERE status<>'Deleted' ORDER BY id DESC LIMIT %v OFFSET %v", TodoItem{}.TableName(), pageLimit, pageOffset)).Scan(data).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go Line:202 ID:7b3ee4, Truy vấn không thành công",
			})
			return
		}
		context.JSON(http.StatusOK, gin.H{
			"data":   data,
			"paging": paging,
		})
	}
}
func DeleteItem(db *gorm.DB) func(*gin.Context) {
	return func(context *gin.Context) {
		id, err := strconv.Atoi(context.Param("id"))
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go Line:218 ID:7fdec9, Id truyền lên không hợp lệ",
			})
			return
		}
		if err := db.Exec(fmt.Sprintf("UPDATE %v SET Status='Deleted' WHERE id=%v", TodoItem{}.TableName(), id)).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go Line:225 ID:c0d632",
			})
			return
		}
		context.JSON(http.StatusOK, gin.H{
			"deleted": true,
		})
	}

}
