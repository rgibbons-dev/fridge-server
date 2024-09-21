package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)

var db *sql.DB

func initDb() (*sql.DB, error) {
	database, err := sql.Open("sqlite3", "./fridge.db")
	if err != nil {
		log.Fatal(err)
	}

	stmt := "create table if not exists links (id integer not null primary key, url text);"

	_, err = database.Exec(stmt)
	if err != nil {
		log.Printf("%q, %s\n", err, stmt)
		return nil, err
	}

	return database, err
}

func middleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

type link struct {
	ID  int    `json:"id"`
	Url string `json:"url"`
}

func getLinks(c *gin.Context) {
	v, ok := c.Get("db")
	if !ok {
		return
	}
	db := v.(*sql.DB)

	rows, err := db.Query("select id, url from links")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var links []link
	for rows.Next() {
		var id int
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		links = append(links, link{ID: id, Url: url})
	}

	c.JSON(http.StatusOK, links)
}

func addLinks(c *gin.Context) {
	v, ok := c.Get("db")
	if !ok {
		return
	}
	db := v.(*sql.DB)

	var json struct {
		Url string `json:"url"`
	}

	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("insert into links (url) values (?)", json.Url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "Link created"})
}

func removeLink(c *gin.Context) {
	v, ok := c.Get("db")
	if !ok {
		return
	}

	db := v.(*sql.DB)
	var json struct {
		Url string `json:"url"`
	}

	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("delete from links where url = ?", json.Url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "Link deleted"})

}

func main() {
	var err error
	db, err = initDb()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	router := gin.Default()
	router.Use(middleware(db))
	router.GET("/links", getLinks)
	router.DELETE("/links", removeLink)
	router.POST("/links", addLinks)
	router.Run(":8081")
}
