package test

import (
    "github.com/gofiber/fiber/v2"
    goWebFramework "github.com/thewindear/go-web-framework"
    "github.com/thewindear/go-web-framework/config"
    "go.uber.org/zap"
    "gorm.io/gorm"
    "log"
    "testing"
)

type Blog struct {
    ID     uint   `gorm:"id" json:"ID"`
    ShopId uint   `gorm:"shop_id" json:"shopId"`
    Title  string `gorm:"title" json:"title"`
    Images string `gorm:"images" json:"images"`
}

func (im Blog) TableName() string {
    return "tb_blog"
}

type UserService struct {
    *goWebFramework.SvcContext
}

func (im UserService) GetBlog() *Blog {
    var blog Blog
    if im.DB().Model(&blog).Where("id = 4").First(&blog).Error == gorm.ErrRecordNotFound {
        return nil
    }
    im.Log().Info("blog info ", zap.String("title", blog.Title))
    return &blog
}

func TestInitCfg(t *testing.T) {
    c, err := goWebFramework.DefaultInitCfg("../configs/config.template.yaml")
    if err != nil {
        t.Fatal(err)
    } else {
        t.Log(c)
    }
}

func TestFramework(t *testing.T) {
    framework, err := goWebFramework.NewFramework("../configs/config.template.yaml", nil)
    if err != nil {
        t.Fatal(err)
    } else {
        framework.SetRouter(func(app *fiber.App) {
            app.Get("/", func(ctx *fiber.Ctx) error {
                userService := &UserService{goWebFramework.NewDefaultSvcContext(ctx.Context(), framework)}
                blog := userService.GetBlog()
                if blog == nil {
                    return ctx.SendStatus(404)
                }
                return ctx.JSON(blog)
            })
        })
        framework.Run()
    }
}

type CustomConfig struct {
    Username         string `json:"username"`
    config.Framework `json:"framework"`
}

func TestCustomConfig(t *testing.T) {
    var customConfig CustomConfig
    _ = goWebFramework.InitCfg[CustomConfig]("../configs/config.template.yaml", &customConfig)
    log.Println(customConfig)
}