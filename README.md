# rrcrud
CRUD example application for Red-River code test

Application may work as API or static website (no Javascript).
It's a GoLang application without modules, so you need to build it etc.

## Build
If you already know, how to do it, skip the following, which is just a
summary of what you can read from the [Go documentation](https://golang.org/cmd/go/#hdr-Compile_packages_and_dependencies) 

 For example, from the directory were `main.go` is: 
 `go build -x -o /tmp/rrcrud` 
 
## Moving template files
The static website requires two template files; one for the root page `index.tmpl` and one for the error page `error.tmpl`. It will not start if they are missing or are invalid.

Two such files are provided in the `templates` directory; you can just copy them to wherever you will indicate the program to look for them at launch with
the `-templates` option.

## Bolt database
The application uses a "permanent" storage in the form of a BoldDb database file (`members.db`). The directory for such file should be specified through the `-bolddb` option.

## Options

Available options can be listed with `-h`

Example:
```
./rrcrud -h                                                                
Usage of ./rrcrud:
  -boltdb string
    	directory for the members.db bolt database (default "/tmp/")
  -noapi
    	do no route the API endpoints
  -nostatic
    	do not route the static endpoints
  -port string
    	listening port (default "8080")
  -templates string
    	directory with the .tmpl template files (default "./templates/")
```


