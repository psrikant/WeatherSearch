# WeatherSearch

## Software Coding Challenge


### Installing

Running the Weather Search program:
(The instructions are specific to unix machine and I have placed the 'WeatherSearch' directory in the home directory)

1) Clone 'WeatherSearch' repository into the home directory

```
cd $HOME
git clone https://github.com/psrikant/WeatherSearch.git
```

2) The following environment variables need to be set:

```
export GOROOT=/usr/local/go
export GOPATH=$HOME/WeatherSearch
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
```

3) Install packages:

```
cd $HOME/WeatherSearch
go get github.com/gorilla/securecookie
go get github.com/mattn/go-sqlite3
go get golang.org/x/crypto/bcrypt
```

4) Change the directory to the server folder:

```
cd src/server
```

5) Run the main.go:

```
go run main.go
```

6) Open the following URL in the browser http://localhost:8081/

7) To display all the user's session history within the command line (Challenge 5), run:

```
go run main.go --display=history
```