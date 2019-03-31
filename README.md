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
The application uses a "permanent" storage in the form of a BoldDb database file (`members.db`). The directory for such file should be specified through the `-boltdb` option.

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
## Static site
Root page is at `/` on the specified port.

## API
The API paths are under the `/api/` root:
* `/api/members` (GET) will return a list of all the members.
* `/api/member/` (POST) will create a new member with the name as specified in the "name" field of the payload. 
* `/api/member/{id}` (GET) will return a list containing only the member with the given ID.
* `/api/member/{id}` (PUT) will replace the name of the member with the one provided in the "name" field on the JSON payload.
* `/api/member/{id}` (DELETE) will removed the member with the given ID.

The JSON response payload may either include a "members" element:

```json
{
    "members":{
      "0e044154-5042-42f9-8d9c-ed03bde85360":{
        "id":"0e044154-5042-42f9-8d9c-ed03bde85360","name":"Natty Bumppo"
      },
      "341533a5-94f6-46d1-83d5-d82365daa17f":{
        "id":"341533a5-94f6-46d1-83d5-d82365daa17f","name":"Luther Blissett"
      }
    },
    "timestamp":"2019-03-31T10:17:50.487129827Z"
}
```

Or an "error":

```json
{
  "error":{
    "code":404,
    "message":"no member with ID 341533a5-94f6-46d1-83d5-d82365daa17g"
  },
  "timestamp":"2019-03-31T10:21:36.015167131Z"
}
```

The "timestamp" element should always be present, if a payload is given (as is not the case with DELETE operations).

As for request payloads:
```json
{"name":"Luther Blissett"}
```
is all you need (everything else is ignored).
