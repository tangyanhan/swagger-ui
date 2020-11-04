package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Repo repo
type Repo struct {
	Name     string
	Branches []Branch
}

// Branch branch
type Branch struct {
	Name        string
	Description string
	Link        string
}

func main() {
	var (
		indexFile   string
		dbPath      string
		externalURL string
	)
	flag.StringVar(&indexFile, "index", "index.html", "index.html to be parsed")
	flag.StringVar(&dbPath, "db", "db.sqlite3", "path to create or load sqlite3 database")
	flag.StringVar(&externalURL, "external", "", "external url to be visited")
	flag.Parse()
	content, err := ioutil.ReadFile(indexFile)
	if err != nil {
		log.Fatalf("failed to load content of %s:%v", indexFile, err)
	}
	tmpl, err := template.New("index").Parse(string(content))
	if err != nil {
		log.Fatalf("failed to parse %s as template:%v", indexFile, err)
	}
	g := gin.New()
	g.Static("/ui/", "./ui")

	db := NewGorm(dbPath)
	g.GET("/swagger", func(c *gin.Context) {
		repo := c.Query("repo")
		branch := c.Query("branch")
		doc, err := GetByRepoBranch(db, repo, branch)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				c.AbortWithStatus(404)
				return
			}
			c.AbortWithError(500, err)
			return
		}
		c.Writer.WriteString(doc.Content)
	})
	g.GET("/upload", func(c *gin.Context) {
		c.File("./upload.html")
	})
	g.POST("/swagger", func(c *gin.Context) {
		repo := c.PostForm("repo")
		branch := c.PostForm("branch")
		description := c.PostForm("description")
		swaggerFile, err := c.FormFile("content")
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		f, err := swaggerFile.Open()
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		content, err := ioutil.ReadAll(f)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}
		doc, err := GetByRepoBranch(db, repo, branch)
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				doc = &Doc{
					Repo:   repo,
					Branch: branch,
				}
			} else {
				c.AbortWithError(500, err)
				return
			}
		}
		doc.Description = description
		doc.Content = string(content)
		if err := CreateDoc(db, doc); err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.Status(200)
		c.Writer.WriteString(externalURL + "/?repo=" + repo + "&branch=" + branch)
	})
	g.GET("/", func(c *gin.Context) {
		all, err := List(db)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		data := struct {
			Repos   map[string][]Branch
			DocLink string
		}{
			Repos:   make(map[string][]Branch),
			DocLink: "/swagger?repo=" + c.Query("repo") + "&branch=" + c.Query("branch"),
		}
		for _, v := range all {
			orig := data.Repos[v.Repo]
			orig = append(orig, Branch{
				Name:        v.Branch,
				Description: v.Description,
				Link:        "/?repo=" + v.Repo + "&branch=" + v.Branch,
			})
			data.Repos[v.Repo] = orig
		}

		if err := tmpl.Execute(c.Writer, &data); err != nil {
			raw, _ := json.Marshal(&data)
			errStr := fmt.Sprintf("Failed to execute template with data %s:%v", string(raw), err)
			c.Writer.WriteString(errStr)
		}
	})
	if err := g.Run(":8080"); err != nil {
		log.Fatalf("failed to run server:%v", err)
	}
}
