## Go programming competency evaluation
### Task: Create a simple REST API in Go
##### API requirements: 

#### Endpoint 1: /items
GET: Returns a JSON list of all current "items" (you define the Item struct). Initially, this list may be empty.


Parameters:
- sort: allows you to specify the sort order of the response.
    - id: (default) - sort results by their ID
    - timestamp: sort results by the date they were created

POST: Adds a new Item to the in-memory data store (no persistence needed).
name is a required field
id should be generated if not provided. It needs to be in the correct format and not duplicate an existing id.

#### Endpoint 2: /items/{id}
GET: Gets a specific Item by its ID, if it exists.

Parameters:
- format: allows you to specify the output format of the response.
    - json (default)
    - yaml

DELETE: Deletes the specified Item matching the id parameter

**Data Considerations**: The Item struct should have at least:
- id: string field in GUID format
- name: string field.
- timestamp: date field

**Code Requirements**
- Use Go's standard library (net/http) to handle HTTP requests.
- Consider using other relevant libraries, such as a JSON encoding package, where appropriate.
- Implement necessary error handling (404 - Not Found, etc.).
- Your code should have suitable test coverage.
- The use of generative AI is permitted to generate base code.

**Evaluation Criteria**
- HTTP Handling: Correct use of request methods (GET, POST, DELETE) and proper handling of routes.
- Data Handling: Appropriate in-memory data structure to store items and logic for adding, retrieving, and deleting items.
- JSON/YAML: Correct marshaling/unmarshaling of items into the requested format.
- Error Handling: Meaningful HTTP status codes and error messages in response to user actions.
- Code Organization: Clean separation of concerns, meaningful function/struct names.
- Input Validation: Check for valid item data during POST.
