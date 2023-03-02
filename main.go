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
	// ID string `json:"id"`
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

// post represents the data of a blog post.
// type post struct {
//     ID string `json:"id"`
//     Title string `json:"title"`
//     Author string `json:"author"`
//     Content string `json:"content"`
// }

// posts slice to seed post data.
// var posts = []post{
//     {ID: "1", Title: "Hello!", Author: "me", Content: "# Hello!<br />Welcome to stream!"},
//     {ID: "2", Title: "Intro", Author: "me", Content: "# Intro<br /> This is a simple, single stream blog built with Go and Astro."},
//     {ID: "3", Title: "Design Choices", Author: "me", Content: "# Design Choices<br />Single author, no char limit but built for short bursts, bad auth, potentially productionizable."},
//     {ID: "4", Title: "Delete Me!", Author: "me", Content: "# Delete Me!<br />To test the API, delete this post and modify the others."},
//     {ID: "5", Title: "The new fad", Author: "me", Content: "# The new fad<br />The new fad on the web is to build your own blog, so here is mine"},
// }


// var db, db_connection_err = gorm.Open(sqlite.Open("stream_data.db"), &gorm.Config{})

// var post Post

// getPosts responds with the list of all posts as JSON.
func getPosts(c *gin.Context) {
	db.Order("id desc").Find(&posts)
	c.IndentedJSON(http.StatusOK, posts)
}

// getRSS responds with the list of all posts as JSON.
func getRSS(c *gin.Context) {
	// https://stackoverflow.com/questions/68836237/how-to-get-full-server-url-from-any-endpoint-handler-in-gin
	// https://blog.canopas.com/how-to-create-rss-feeds-in-golang-43a99fa302e9
	feed := &feeds.Feed{
		Title:       os.Getenv("RSS_FEED_TITLE"),
		Link:        &feeds.Link{Href: c.Request.Host + "/feed"},
		Description: os.Getenv("RSS_FEED_DESCRIPTION"),
		Author:      &feeds.Author{Name: os.Getenv("RSS_FEED_AUTHOR_NAME"), Email: os.Getenv("RSS_FEED_AUTHOR_EMAIL")},
		// Created:     time.Now(),
	}
	
	var feedItems []*feeds.Item
	
	for _, post := range posts {
	// for i := 0; i < len(posts); i++ {  
			// item := posts[i]
			// fmt.Println("post.ID - ")
			// fmt.Println(post.ID)
			// https://stackoverflow.com/questions/57187889/how-to-convert-uint-type-into-string-type-in-golang
			feedItems = append(feedItems,
				&feeds.Item{
					Id:			 fmt.Sprintf("%s/posts/%s", c.Request.Host, strconv.FormatUint(uint64(post.ID), 10)),
					Title:		 post.Title,
					Link:		 &feeds.Link{Href: c.Request.Host + "/posts/" + strconv.FormatUint(uint64(post.ID), 10)},
					Description: post.ContentHTML,
					Created:	 post.CreatedAt,
					// Author:		 &feeds.Author{Name: "Nitin Khanna", Email: "mail@nitinkhanna.com"},
			})
			// fmt.Println("feedItems - ")
			// fmt.Println(feedItems)
	}
	// fmt.Println("feedItems - ")
	// fmt.Println(feedItems)

	feed.Items = feedItems
	
	// rssFeed := (&feeds.Rss{Feed: feed}).RssFeed()
	// xmlRssFeeds := rssFeed.FeedXml()

	// fmt.Println("feed -")
	// fmt.Println(feed)

	rss, err := feed.ToRss()
	if err != nil {
	    c.IndentedJSON(http.StatusNotFound, gin.H{"message": "feed not found"})
	    return
	}

	// c.XML(http.StatusOK, feed)
	c.Writer.WriteString(rss)
}

// getPostByID locates the post whose ID value matches the id
// parameter sent by the client, then returns that post as a response.
func getPostByID(c *gin.Context) {
	var post Post
	id := c.Param("id")
	// db.First(&post, id) 

	result := db.First(&post, id)
	// result.RowsAffected // returns count of records found
	// result.Error        // returns error or nil
	// check error ErrRecordNotFound
	// errors.Is(result.Error, gorm.ErrRecordNotFound)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// handle record not found
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "post not found"})
	}

	c.IndentedJSON(http.StatusOK, post)
	// return

	// Loop over the list of posts, looking for
	// an post whose ID value matches the parameter.
	// for _,a := range posts {
	//     if a.ID == id {
	//         c.IndentedJSON(http.StatusOK, a)
	//         return
	//     }
	// }
	// c.IndentedJSON(http.StatusNotFound, gin.H{"message": "post not found"})
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
	// posts = append(posts, newPost)
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
	// for i, a := range posts {
	//     if a.ID == id {
	//         posts = append(posts[:i], posts[i+1:]...)
	//         c.IndentedJSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
	//         return
	//     }
	// }
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
	// fmt.Println("ID from URL is - ")
	// fmt.Println(id)
	result := db.First(&post, id)

	// fmt.Println("post found is - ")	
	// fmt.Println(post)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// handle record not found
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "post not found"})
	}
	Content_md := []byte(newPost.Content)
	html := markdown.ToHTML(Content_md, nil, nil)
	db.Model(&post).Updates(Post{Content: newPost.Content, ContentHTML: string(html)})



	// for i, a := range posts {
	// 	if a.ID == id {
	// 		posts[i] = newPost
	// 		c.IndentedJSON(http.StatusOK, newPost)
	// 		return
	// 	}
	// }
	// c.IndentedJSON(http.StatusNotFound, gin.H{"message": "post not found"})
}


// getLogin validates the login and redirects back to index
func getLogin(c *gin.Context) {
	user := c.MustGet(gin.AuthUserKey).(string)
	// c.Request.URL.Path = "/"
	if secret, ok := secrets[user].(string); ok {
		// c.JSON(http.StatusOK, gin.H{"user": user, "secret": secret})
		current_secret = secret
		authorized = true
	} else {
		// c.JSON(http.StatusOK, gin.H{"user": user, "secret": "NO SECRET :("})
		current_secret = ""
		authorized = false
	}
	c.Redirect(http.StatusFound, "/")
	// r.HandleContext(c)
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
	// r.JSON(c.Writer, http.StatusOK, map[string]string{"welcome": "This is rendered JSON!"})
	// c.JSON(http.StatusOK, gin.H{
	//     "message": "Hello World! We are here!",
	// })
}

func (p *Post) AfterFind(tx *gorm.DB) (err error) {
	// fmt.Println(p.CreatedAt)
	// fmt.Printf("p.CreatedAt = %T\n", p.CreatedAt)
	// fmt.Println("Unix format:", p.CreatedAt.Format(time.UnixDate))
	// fmt.Println("time.RFC3339 format:", p.CreatedAt.Format(time.RFC3339))
	// t, err := time.Parse(time.UnixDate, p.CreatedAt.Format(time.UnixDate))
	// fmt.Println(t)
	t1 := time.Date(p.CreatedAt.Year(), p.CreatedAt.Month(), p.CreatedAt.Day(), p.CreatedAt.Hour(), p.CreatedAt.Minute(), p.CreatedAt.Second(), p.CreatedAt.Nanosecond(), p.CreatedAt.Location())
	// fmt.Println(t1)
	// p.CreatedDateFormat = p.CreatedAt.Format(time.RFC3339)
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

	// fmt.Printf("db = %T\n", db)
	// fmt.Printf("secrets = %T\n", secrets)
	
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
	// authorized.POST("/login", getLoginHTML)

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