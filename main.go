package main

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ItemStatus int

const (
	Doing ItemStatus = iota
	Done
	Deleted
)
const TABLE_NAME = "todo_items"

var allItemStatus = []string{"Doing", "Done", "Deleted"}

func (item *ItemStatus) String() string {
	return allItemStatus[*item]
}

func parseStr2ItemStatus(s string) (ItemStatus, error) {
	// Chuyển đổi string thành ItemStatus
	for i := range allItemStatus {
		if s == allItemStatus[i] {
			return ItemStatus(i), nil
		}
	}
	return Doing, errors.New("invalid status string")
}

func (item *ItemStatus) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Scan du lieu bi loi: %v", value)
	}
	v, err := parseStr2ItemStatus(string(bytes))
	if err != nil {
		return fmt.Errorf("Scan du lieu bi loi: %v", value)
	}
	*item = v
	return nil
}

func (item *ItemStatus) Value() (driver.Value, error) {
	if item == nil {
		return nil, nil
	}
	return item.String(), nil
}

func (item *ItemStatus) MarshalJSON() ([]byte, error) {
	if item == nil {
		return nil, nil
	}
	return []byte(fmt.Sprintf("\"%s\"", item.String())), nil
}

func (item *ItemStatus) UnmarshalJSON(data []byte) error {
	str := strings.ReplaceAll(string(data), "\"", "")
	itemValue, err := parseStr2ItemStatus(str)
	if err != nil {
		return err
	}
	*item = itemValue
	return nil
}

type TodoItem struct {
	Id          int         `json:"id" gorm:"column:id;"`
	Title       string      `json:"title" gorm:"column:title;"`
	Description string      `json:"description" gorm:"column:description;"`
	Status      *ItemStatus `json:"status" gorm:"column:status;"`
	CreatedAt   *time.Time  `json:"created_at" gorm:"column:created_at;"`
	UpdatedAt   *time.Time  `json:"updated_at" gorm:"column:updated_at;"`
}
type TodoItemCreation struct {
	Id          int         `json:"-" gorm:"column:id;"`
	Title       string      `json:"title" gorm:"column:title;"`
	Description string      `json:"description" gorm:"column:description;"`
	Status      *ItemStatus `json:"status" gorm:"column:status;"`
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

func (TodoItem) TableName() string         { return TABLE_NAME }
func (TodoItemCreation) TableName() string { return TABLE_NAME }
func (TodoItemUpdate) TableName() string   { return TABLE_NAME }

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
				"detail": "🎁 main.go ID:30ad16",
			})
			return
		}
		data.Id = id
		if err := db.First(data).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go ID:d3b91f",
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
				"detail": "🎁 main.go ID:061053",
			})
			return
		}
		if err := context.ShouldBindJSON(data); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go ID:802374",
			})
			return
		}
		if err := db.Where("id = ?", id).Updates(data).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go ID:075586, Cập nhật không thành công",
			})
			return
		}
		context.JSON(http.StatusOK, gin.H{
			"updated": true,
		})
	}
}
func ListItem(db *gorm.DB) func(*gin.Context) {
	return func(context *gin.Context) {
		//Parse tu param
		paging := new(Paging)
		if err := context.ShouldBind(paging); err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go ID:a35b47, Dữ liệu truyền lên không đúng định dạng",
			})
			return
		}

		paging.Process()

		if err := db.Raw(fmt.Sprintf("SELECT count(*) FROM %v", TABLE_NAME)).Scan(&paging.Total).Error; err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":  err.Error(),
				"detail": "🎁 main.go ID:dc8aa9, Khong dem duoc",
			})
			return
		}

		data := new([]TodoItem)
		pageLimit := paging.Limit
		pageOffset := (paging.Page - 1) * pageLimit
		if err := db.Raw(fmt.Sprintf("SELECT * FROM %v WHERE status<>'Deleted' ORDER BY id DESC LIMIT %v OFFSET %v", TABLE_NAME, pageLimit, pageOffset)).Scan(data).Error; err != nil {
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
				"detail": "🎁 main.go ID:f63371, Id truyền lên không hợp lệ",
			})
			return
		}
		if err := db.Exec(fmt.Sprintf("UPDATE %v SET Status='Deleted' WHERE id=%v", TABLE_NAME, id)).Error; err != nil {
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
