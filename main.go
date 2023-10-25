package main

import (
	"database/sql"  //import for connecting to the sql database
	"encoding/json" // to accept data in json format
	"fmt"
	"log"
	"net/http" // to handle http requests
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // postgress dependencies
)

// the structure our data would take after being recieved
type ApodData struct {
	Copyright      string `json:"copyright"`
	Date           string `json:"date"`
	Explanation    string `json:"explanation"`
	Hdurl          string `json:"hdurl"`
	MediaType      string `json:"media_type"`
	ServiceVersion string `json:"service_version"`
	Title          string `json:"title"`
	URL            string `json:"url"`
}

func main() {
	// connecting to database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL")) // os.Getenv is for getting the environment variable
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//make API request and retrieve data

	apiURL := "https://api.nasa.gov/planetary/apod?api_key=h8wHbN8UdrOOILRFpKxEVlmu3XQ7lgGTmzL9Iuee"
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var data ApodData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatal(err)
	}
	//fmt.Println(data.Title)

	// create the table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS apoddata (id SERIAL PRIMARY KEY, explanation TEXT, title TEXT, url TEXT)")
	if err != nil {
		log.Fatal(err)
	}
	// Insert data into the PostgreSQL table
	_, err = db.Exec("INSERT INTO apoddata (explanation, title, url) VALUES ($1, $2, $3) RETURNING id", data.Explanation, data.Title, data.URL)
	if err != nil {
		log.Fatal(err)
	}

	//create router
	// for handling different requests
	router := mux.NewRouter()
	router.HandleFunc("/apoddata", getUsers(db)).Methods("GET")       // for handling the http GET request(all users)
	router.HandleFunc("/apoddata/{id}", getUser(db)).Methods("GET")   // This line sets up a new route on the router to handle HTTP GET requests for the path "/users/{id}"
	router.HandleFunc("/apoddata", createUser(db)).Methods("POST")    // for handling the http POST request
	router.HandleFunc("/apoddata{id}", updateUser(db)).Methods("PUT") // for handling the http PUT request(updating the user)
	router.HandleFunc("/apoddata/{id}", deleteUser(db)).Methods("DELETE")

	// start server
	/*the provided code starts an HTTP server on port 8000 and
	sets up a middleware to ensure that the "Content-Type" header
	of the HTTP response is always set to "application/json".
	This middleware is applied to all incoming requests,
	ensuring consistent response formatting.
	the code is also nested in log.fatal so if an error occurs during the http server setup the error can be logged and the program will be exited */
	log.Fatal(http.ListenAndServe(":8000", jsonContentTypeMiddleware(router))) // for handling any error gotten on the requset on the port
	fmt.Println(router.HandleFunc("/apoddata", getUsers(db)).Methods("GET"))
}

// jsonContentTypeMiddleware formats response as JSON for client
func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	// http.Handler func allows  a function with a specific signature to be used as an http.Handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//w http.ResponseWriter: This parameter represents the response writer, allowing you to write the response back to the client.
		//r *http.Request: This parameter represents the HTTP request received from the client.
		w.Header().Set("Content-Type", "application/json")
		/*So, this line of code is setting the "Content-Type" header of the HTTP response to indicate that the response content will be in JSON format. This is commonly used when building APIs to specify the format of the response being sent back to the client.*/
		/*w.Header(): This is a method of http.ResponseWriter that allows us to access
				and manipulate the headers of the HTTP response.
				It returns an http.Header object, which represents the response headers.
		        Set("Content-Type", "application/json"): This is a method of http.Header
				 that sets a specific header key-value pair. In this case, we are setting
				 the "Content-Type" header to "application/json". The "Content-Type" header
				 indicates the media type of the response content, and "application/json" specifies t
				 hat the response will be in JSON format.*/
		next.ServeHTTP(w, r)
	})
}

// get all users
/*The code snippet  fetches all users from a database using an SQL query, scans the results into a slice of User structs, and handles any errors encountered during this process. This can be used to retrieve user data from a database and prepare it for a response in an HTTP handler.*/
func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*The line below executes a SQL query to select all records
		from a table named "USERS" using the provided database connection (db).
		It stores the result in rows, and any error encountered is stored in err.*/
		rows, err := db.Query("SELECT * FROM apoddata")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		users := []ApodData{}
		var u ApodData
		for rows.Next() { // rows.Next() makes loop to continue as long as there are more rows
			/*the line of code below scans the values from the current row of the result set into the provided variables (u.ID, u.Name, and u.Email). If any error occurs during scanning, it logs the error and terminates the program, ensuring that any critical errors are appropriately handled.*/
			if err := rows.Scan(&u.Explanation, &u.Title, &u.URL, &u.Date); err != nil { // rows.Scan() is used to scan the values from the current row of the result set into variables provided as arguments
				log.Fatal(err)
			}
			users = append(users, u) // Appends the current user (u) to the users slice.
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(users) // for returning the users encoded in json
	}
}

// get user  by id

func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r) // from the Gorilla Mux router package to extract variables from the HTTP request URL. In this case, it's used to extract the "id" parameter from the URL.
		id := vars["id"]    //This line retrieves the value of the "id" parameter from the URL, which was extracted in the previous line.

		var u ApodData
		err := db.QueryRow("SELECT * FROM apoddata WHERE id = $1", id).Scan(&u.Explanation, &u.Title, &u.URL)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}
		json.NewEncoder(w).Encode(u)
	}
}

// create user
/*this code defines a handler for creating a user. It expects a JSON representation of a user in the HTTP request body, decodes it into a User struct, inserts the user into the database, retrieves the user's generated ID, and encodes the user (including the generated ID) as a JSON response. If any error occurs during this process, it logs the error and terminates the program.*/
func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u ApodData
		json.NewDecoder(r.Body).Decode(&u) // This line uses json.NewDecoder to decode the JSON data from the HTTP request body (r.Body) into the u variable, which represents a user. The &u is a pointer to u, allowing the decoder to modify its content.
		err := db.QueryRow("INSERT INTO apoddata (description, title, Imageurl) VALUES ($1, $2, $3) RETURNING id", u.Explanation, u.Title, u.URL).Scan(&u.Title)
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode(u)
	}
}

// update user
func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u ApodData
		json.NewDecoder(r.Body).Decode(&u)

		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE apoddata SET description = $1, title = $2, Imageurl = $3, WHERE id = $3", u.Explanation, u.Title, u.Explanation, id)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(u)
	}
}

// delete user
func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r) // extract variables from the HTTP request URL.
		id := vars["id"]

		var u ApodData
		err := db.QueryRow("SELECT * FROM apoddata WHERE id = $1", id).Scan(&u.Explanation, &u.Title, &u.URL)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := db.Exec("DELETE FROM apoddata WHERE id = $1", id)
			if err != nil {
				//todo : fix error handling
				w.WriteHeader(http.StatusNotFound)
				return
			}

			json.NewEncoder(w).Encode("User deleted")
		}
	}
}
