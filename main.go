package main

import (
	"os"
	"fmt"
	"time"
	"errors"
	"strconv"
	"net/http"
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/unrolled/render"
	"github.com/gomarkdown/markdown"
)

type Post struct {
	gorm.Model
	Title string `json:"title"`
	Author string `json:"author"`
	Content string `json:"content"`
	ContentHTML string `json:"contentHTML"`
	CreatedDateFormat string `json:"createdDateFormat"`
}

var stream_username string
var secrets gin.H
var current_secret string
var authorized = false
var db *gorm.DB
var posts []Post

// getPosts responds with the list of all posts as JSON.
func getPosts(c *gin.Context) {
	db.Order("id desc").Find(&posts)
	c.IndentedJSON(http.StatusOK, posts)
}

// getRSS responds with the list of all posts as JSON.
func getRSS(c *gin.Context) {
	feed := &feeds.Feed{
		Title:       os.Getenv("RSS_FEED_TITLE"),
		Link:        &feeds.Link{Href: c.Request.Host + "/feed"},
		Description: os.Getenv("RSS_FEED_DESCRIPTION"),
		Author:      &feeds.Author{Name: os.Getenv("RSS_FEED_AUTHOR_NAME"), Email: os.Getenv("RSS_FEED_AUTHOR_EMAIL")},
	}
	
	var feedItems []*feeds.Item
	
	for _, post := range posts {
		feedItems = append(feedItems,
			&feeds.Item{
				Id:			 fmt.Sprintf("%s/posts/%s", c.Request.Host, strconv.FormatUint(uint64(post.ID), 10)),
				Title:		 post.Title,
				Link:		 &feeds.Link{Href: c.Request.Host + "/posts/" + strconv.FormatUint(uint64(post.ID), 10)},
				Description: post.ContentHTML,
				Created:	 post.CreatedAt,
		})
	}

	feed.Items = feedItems
	
	rss, err := feed.ToRss()
	if err != nil {
	    c.IndentedJSON(http.StatusNotFound, gin.H{"message": "feed not found"})
	    return
	}

	c.Writer.WriteString(rss)
}

// getPostByID locates the post whose ID value matches the id
// parameter sent by the client, then returns that post as a response.
func getPostByID(c *gin.Context) {
	var post Post
	id := c.Param("id")

	result := db.First(&post, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// handle record not found
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "post not found"})
	}

	c.IndentedJSON(http.StatusOK, post)
}

// createPosts adds a new post from JSON received in the request body.
func createPosts(c *gin.Context) {
	var newPost Post

	token_data := c.Request.Header.Get("Token")
	if token_data != current_secret {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "Ya Can't Do That!"})
		return
	}

	// Call BindJSON to bind the received JSON to
	// newPost.
	if err := c.BindJSON(&newPost); err != nil {
		return
	}

	Content_md := []byte(newPost.Content)
	html := markdown.ToHTML(Content_md, nil, nil)

	newPost.ContentHTML = string(html)

	// Add the new post to the slice.
	db.Create(&newPost)
	c.IndentedJSON(http.StatusCreated, newPost)
}

// deletePosts deletes a post by the ID from JSON received in the request body.
func deletePosts(c *gin.Context) {
	var post Post
	token_data := c.Request.Header.Get("Token")
	if token_data != current_secret {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "Ya Can't Do That!"})
		return
	}

	id := c.Param("id")
	db.Delete(&post, id)
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "post deleted successfully"})
}

// updatePosts updates the content of a post by the ID from JSON received in the request body.
func updatePosts(c *gin.Context) {
	var post Post
	var newPost Post

	token_data := c.Request.Header.Get("Token")
	if token_data != current_secret {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "Ya Can't Do That!"})
		return
	}

	// Call BindJSON to bind the received JSON to
	// newPost.
	if err := c.BindJSON(&newPost); err != nil {
		return
	}

	id := c.Param("id")
	result := db.First(&post, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// handle record not found
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "post not found"})
	}
	Content_md := []byte(newPost.Content)
	html := markdown.ToHTML(Content_md, nil, nil)
	db.Model(&post).Updates(Post{Content: newPost.Content, ContentHTML: string(html)})
}


// getLogin validates the login and redirects back to index
func getLogin(c *gin.Context) {
	user := c.MustGet(gin.AuthUserKey).(string)
	if secret, ok := secrets[user].(string); ok {
		current_secret = secret
		authorized = true
	} else {
		current_secret = ""
		authorized = false
	}
	c.Redirect(http.StatusFound, "/")
}


// getLogout invalidates the login and redirects back to index
func getLogout(c *gin.Context) {
	authorized = false
	c.Redirect(http.StatusFound, "/")
}


// getIndexHTML responds with HTML for the index page
func getIndexHTML(c *gin.Context) {
	r := render.New(render.Options{
		IndentJSON: true,
		IsDevelopment: true,
		UnEscapeHTML: true,
	})
	db.Order("id desc").Find(&posts)
	if authorized {
		r.HTML(c.Writer, http.StatusOK, "index", gin.H{
			"posts": posts,
			"authorized": authorized,
			"current_secret": current_secret,
		})
	} else {
		r.HTML(c.Writer, http.StatusOK, "index", gin.H{
			"posts": posts,
			"authorized": authorized,
			"current_secret": "",
		})
	}
}

func (p *Post) AfterFind(tx *gorm.DB) (err error) {
	t1 := time.Date(p.CreatedAt.Year(), p.CreatedAt.Month(), p.CreatedAt.Day(), p.CreatedAt.Hour(), p.CreatedAt.Minute(), p.CreatedAt.Second(), p.CreatedAt.Nanosecond(), p.CreatedAt.Location())
	p.CreatedDateFormat = t1.Format(time.RFC3339)
	return
}


func main() {
	err := godotenv.Load()
	if err != nil {
	    panic("Error loading .env file. Please add a .env file in the current working directory first.")
	}

	secrets = gin.H{
		os.Getenv("STREAM_USER"): os.Getenv("STREAM_SECRET"),
	}
	current_secret = os.Getenv("STREAM_SECRET")

	db, err = gorm.Open(sqlite.Open("stream_data.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	
	// Migrate the schema
	db.AutoMigrate(&Post{})
	// StatusCreated
	db.FirstOrCreate(&Post{Author: "me", Title: "Hello!", Content: "Welcome to your stream!", ContentHTML: "<p>Welcome to your stream!</p>"}, Post{Title: "Hello!"})
	db.FirstOrCreate(&Post{Author: "me", Title: "Intro", Content: "This is a simple, single stream blog built with Go and tailwind.", ContentHTML: "<p>This is a simple, single stream blog built with Go and tailwind.</p>"}, Post{Title: "Intro"})
	db.FirstOrCreate(&Post{Author: "me", Title: "Design Choices", Content: "Single author, no char limit but built for short bursts. Bad auth; potentially productionizable.", ContentHTML: "<p>Single author, no char limit but built for short bursts. Bad auth; potentially productionizable.</p>"}, Post{Title: "Design Choices"})
	db.FirstOrCreate(&Post{Author: "me", Title: "Delete Me!", Content: "To test the API, delete this post and modify the others.", ContentHTML: "<p>To test the API, delete this post and modify the others.</p>"}, Post{Title: "Delete Me!"})
	db.FirstOrCreate(&Post{Author: "me", Title: "The new fad", Content: "The new fad on the web is to build your own blog, so here is mine", ContentHTML: "The new fad on the web is to build your own blog, so here is mine"}, Post{Title: "The new fad"})

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// Group using gin.BasicAuth() middleware
	// gin.Accounts is a shortcut for map[string]string
	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		os.Getenv("STREAM_USER"): os.Getenv("STREAM_PASSWORD"),
	}))

	authorized.GET("/login", getLogin)
	authorized.GET("/logout", getLogout)

	router.Static("/assets", "./assets")
	router.GET("/", getIndexHTML)
	router.GET("/posts", getPosts)
	router.GET("/feed", getRSS)
	router.GET("/posts/:id", getPostByID)
	router.POST("/posts", createPosts)
	router.PUT("/posts/:id", updatePosts)
	router.DELETE("/posts/:id", deletePosts)

	db.Order("id desc").Find(&posts)

	router.Run("localhost:8080")
}