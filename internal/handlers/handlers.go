// Package handlers
package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/WWoi/web-parcer/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/imbecility/go-fake-useragent/useragent"
)

const SearchURL = "https://u-card.wb.ru/cards/v4/detail?appType=1&curr=rub&dest=123589415&spp=30&ab_testing=false&ab_testing=false&lang=ru&nm="

func SearchByArticul(c *gin.Context) {
	articul := c.Param("articul")

	query := SearchURL + articul

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, query, nil)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to create request:" + err.Error(),
		})
		return
	}

	// set headers
	gen, _ := useragent.NewGenerator()
	req.Header.Set("User-Agent", gen.Get())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")
	req.Header.Set("Origin", "https://www.wildberries.ru")
	req.Header.Set("Referer", "https://www.wildberries.ru/")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "request failed:" + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "failed to read body:" + err.Error(),
		})
		return
	}

	// parsing to struct
	var ResponseBody models.ResponseBody
	if err := json.Unmarshal(body, &ResponseBody); err != nil {
		c.JSON(500, gin.H{
			"error": "failed to parse JSON:" + err.Error(),
		})
		return
	}

	if ResponseBody.Products[0].RcID == 0 {
		fmt.Println("Product is not original")
	} else {
		fmt.Println("Legit check succsessfuly passed")
	}

	fmt.Println(ResponseBody)

	c.JSON(200, gin.H{
		"message": "All right!",
		"body":    ResponseBody,
	})
}
