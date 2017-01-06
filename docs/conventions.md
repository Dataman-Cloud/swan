
## Commit Message (Better to follow)

* Separate subject from body with a blank line
* Limit the subject line to 50 characters
* Capitalize the subject line
* Do not end the subject line with a period
* Use the imperative mood in the subject line
* Wrap the body at 72 characters
* Use the body to explain what and why vs. how

Ref:
[http://chris.beams.io/posts/git-commit/#separate](http://chris.beams.io/posts/git-commit/#separate)


## API

### Verbs on Resources

API resources should use the traditional REST pattern:

- GET `/<resourceNamePlural>`
    - Retrieve a list of type <resourceName>, e.g. GET /apps returns a list of Apps.
- POST `/<resourceNamePlural>`
    - Create a new resource from the JSON object provided by the client.
- GET `/<resourceNamePlural>/<name>`
    - Retrieves a single resource with the given name, e.g. GET /apps/nginx returns a App named 'nginx'.
- DELETE `/<resourceNamePlural>/<name>`
    - Delete the single resource with the given name.
- PUT `/<resourceNamePlural>/<name>`
    - Update or create the resource with the given name with the JSON object provided by the client.
- PATCH `/<resourceNamePlural>/<name>`
    - Selectively modify the specified fields of the resource.

We do have some APIs that has action in the path with Method `POST` or `PUT`, also we may use `DELETE`
to delete a collection of resources, but we should do our best to make our APIs as most RESTful as possible.

Ref:
[K8s API Conventions](https://github.com/kubernetes/kubernetes/blob/master/docs/devel/api-conventions.md#verbs-on-resources)


### HTTP Status codes

GET list:
- 200 Okay
- 400 Bad request
- 401 Unauthorized
- 403 Forbidden
- 500 Internal server error

GET with ID in path:
- 200 Okay
- 400 Bad request
- 401 Unauthorized
- 403 Forbidden
- 404 Not found(resource ID is invalid, not existed)
- 500 Internal server error

POST as Create:
- 201 Created
- 400 Bad request
- 401 Unauthorized
- 403 Forbidden
- 404 Not found(resource ID is invalid, not existed)
- 409 Conflict
- 500 Internal server error

POST as action:
- 200 Okay
- 400 Bad request
- 401 Unauthorized
- 403 Forbidden
- 404 Not found(resource ID is invalid, not existed)
- 500 Internal server error

DELETE:
- 200 Okay (if we provide response body)
- 204 No Content (if empty response)
- 401 Unauthorized
- 400 Bad request
- 403 Forbidden
- 404 Not found(resource ID is invalid, not existed)
- 500 Internal server error

PUT as Update:
- 200 Okay
- 401 Unauthorized
- 400 Bad request (only if with user input: e.g. query params)
- 403 Forbidden
- 404 Not found(resource ID is invalid, not existed)
- 500 Internal server error

PATCH as Update:
- 200 Okay
- 401 Unauthorized
- 400 Bad request (only if with user input: e.g. query params)
- 403 Forbidden
- 404 Not found(resource ID is invalid, not existed)
- 500 Internal server error
