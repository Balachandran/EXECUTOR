package main

//curl -X POST http://192.168.1.151:9001/calculatesum  -H 'Content-Type: application/json'  -d "{\"a\":100,\"b\":200}"


import (
	 
	"fmt"
	"net/http"
	"os"
	"time" 
	//"io/ioutil"
    "log"
    "encoding/json"
	"strconv"
	"database/sql"
	"github.com/gorilla/mux"
	 _ "github.com/go-sql-driver/mysql"
	 
)
var Database *sql.DB

type AddRequest struct {
	A int `json:"a"`
	B int `json:"b"`
} 

func  calculatesum(rw http.ResponseWriter, r *http.Request) {
 
 
 log.Println("Calling calculate sum") 

	
  	var request AddRequest
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Perform addition
	c := request.A + request.B
	
	stmt, err := Database.Prepare("INSERT INTO temp (a,b,c) VALUES(?,?,?)")
	if err != nil {
	  	log.Println(err.Error())
	    json.NewEncoder(rw).Encode(map[string]string{"result": err.Error()}) 
	    return
	} 
	  
   //fmt.Println(request.A)
   //fmt.Println(request.B)  
   
  _, err = stmt.Exec(request.A,request.B)
  if err != nil {
  	log.Println(err.Error())
    json.NewEncoder(rw).Encode(map[string]string{"result": err.Error()}) 
    return
  }
 json.NewEncoder(rw).Encode(map[string]string{"result": strconv.Itoa(c)}) 
} 



func main() {
	
	
	currentTime := time.Now()
   f, err := os.OpenFile("/home/bala/goprojects/POC/API1/log.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
   if err != nil {
	    log.Fatalf("error opening file: %v", err)
   }
   defer f.Close()
	
   log.SetOutput(f)
   log.Println("Starting the REST API Server ",currentTime)
    
    db, err := sql.Open("mysql", "E2EUSER:Automation$1234@tcp(192.168.1.151:3306)/E2E")
    if err != nil {
    	log.Println("Error opening SQL Connection",err.Error())
        panic(err.Error())
        return
    }    
    Database = db
    router := mux.NewRouter()    
    
    
    router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for preflight response
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	})
    
    //Route all the application related REST API 
    router.HandleFunc("/calculatesum",calculatesum).Methods("POST");  
    
	fmt.Println("Starting Server on port 9001......",currentTime)
    log.Fatal(http.ListenAndServe(":9001", router))  
    defer db.Close()
}