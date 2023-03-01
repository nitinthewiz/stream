# stream
An experimental stream of conscious blog written in Go.

# What is it?
A simple "liveblog", my favorite kind of blog, written in Go.

# Why did you build it?
I wanted to learn Go. I was told to build a REST API in Go in order to learn something interesting. But what's a REST API got to do? Why not serve a blog? That'd be cool, right?

# Does it support all endpoints of a REST API?
Yes, but. While I do support all CRUD ops, the web interface doesn't allow DELETE right now. It only allows CREATE, READ, and UPDATE. Maybe at some point I'll add delete. Should be simple enough.

# What all libaries are you using?
- The database is sqlite and is accesssed through gorm
- The app and API are served using Gin
- Environment variables, including the username and password are accessed using joho/godotenv
- The RSS feed is built using gorilla/feeds (can't have a blog without an RSS feed!)
- The index template is rendered using unrolled/render
- The blog supports markdown input, and converts that into HTML using gomarkdown/markdown
- Tailwindcss and momentJS provide some much needed frontend niceness.

# How do I go about using this?
- Install Go v1.19.5 or above at the location where you want to run this (say, your Ubuntu VPS)
- Grab the entire repo and build using `go build`
- This will create the `stream` binary.
- The most important files to run the server are - 
	- the `stream` binary 
	- .env_example file
	- the assets folder 
	- the templates folder
- rename the .env_example file to .env and fill in the variables there - 
	- STREAM_USER 		- make this the user you will login to your blog with
	- STREAM_PASSWORD   - make this the password you will login to your blog with
	- STREAM_SECRET     - make this a long and secretive string that the app will pass to the browser after you've logged in. Once the browser has it, you're considered logged in, and can create or edit posts. You can also use this secret to do CRUD operations using the API. Just specify the secret in the header as "Token" and you'll be considered authenticated.
	- RSS_FEED_TITLE=Your Fancy Blog Name
	- RSS_FEED_DESCRIPTION=A description of the blog
	- RSS_FEED_AUTHOR_NAME=Your Name
	- RSS_FEED_AUTHOR_EMAIL=Your Email ID
- Now run the binary!
- If you're using nginx to host, setup the vhost as proxy_pass
- If you're using ufw, make sure to open up port 8080
- ???
- $$$

# This doesn't seem very secure.
Yes. Here be dragons. Use at your own peril. Etc etc. This is more of a fun learning project than a full blown web app.

# How do I back up my posts?
Periodically copy over the stream_data.db file to a nice backup place, like dropbox.

# OMG this is cool!
Thanks!