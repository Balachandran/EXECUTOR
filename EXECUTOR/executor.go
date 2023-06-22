package main

//curl -X POST http://192.168.1.151:9001/calculatesum  -H 'Content-Type: application/json'  -d "{\"a\":100,\"b\":200}"
//curl -X POST http://192.168.1.151:9003/submit  -H 'Content-Type: application/json'  -d "{\"profileid\":1,\"username\":\"pulliybx\"}"

import (
	 
	"fmt"
	"net/http" 
	"time" 
	//"io/ioutil"
    "log"
    "encoding/json"
	"strconv"
	"database/sql"
	"github.com/gorilla/mux"
	 _ "github.com/go-sql-driver/mysql"
	 "github.com/google/uuid"
	"os"
	"os/exec"
	//"path/filepath"
	 
)
var Database *sql.DB

type Profile  struct{ 
 
  Id		   int
  AppId        int            
  Name         string
  RootFolder   string
  Createdby    string
  Createdon    string    
  Exectype     string 
  Executedon   string
  Giturl       string
  Gitkey  	   string
  Notes		   string
  EmailId	   string
}

type Testcase struct{
	Id int
	RelativePath string
	Executable string
}


type JobRequest struct {
	Profileid int  
	Username string  
} 



var jobQueue []JobRequest


type ExecRequest struct {
	Profileid int `json:"profileid"` 
	Username string `json:"username"` 
} 

func PollQueue() {
	for {
		if len(jobQueue) > 0 {
			// Process the job 
			fmt.Println("Processing Profile:", jobQueue[0].Profileid," username ",jobQueue[0].Username)
			executeprofile(jobQueue[0].Profileid,jobQueue[0].Username)	 
			// Remove the processed job from the queue
			jobQueue = jobQueue[1:]
		}

		// Sleep for 1 second before checking the queue again
		time.Sleep(1 * time.Second)
	}
}

func executeprofile(profileId int,username string){
	
	log.Println("Executing Profile ",profileId, " for username ",username) 
	var row Profile 
	query := "SELECT Id,AppId,Name,RootFolder,Createdby,Createdon,Exectype,COALESCE(Executedon,'') as Executedon,Giturl,Gitkey from profiles WHERE Id=? and Createdby=?"
	err := Database.QueryRow(query, profileId,username).Scan(&row.Id,&row.AppId,&row.Name,&row.RootFolder,&row.Createdby,&row.Createdon,&row.Exectype,&row.Executedon,&row.Giturl,&row.Gitkey)
    if err != nil {
	    log.Println(err.Error())
        //http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}   
    guid := uuid.New() 
	guidStr := guid.String()
	
    log.Println("Creating build folder ",guidStr)  
	direrr := os.Mkdir(guidStr, 0755)
	if direrr != nil { 
	    log.Println(direrr.Error())
        //http.Error(rw, direrr.Error(), http.StatusBadRequest)
		return
	} 
	
	folderPath := guidStr
	repoURL :=	row.Giturl   
	branch := row.Name
	log.Println("Extracting artifacts from GIT for ",branch)
	cmd := exec.Command("git", "clone", "-b", branch, repoURL, folderPath)
	cmd.Dir = "./"
	execerr := cmd.Run()
	if execerr != nil {
		log.Println("Error Executing Profile ",execerr.Error())
        //http.Error(rw, execerr.Error(), http.StatusBadRequest)
        return
	}
	log.Println("Downloaded  artifacts from GIT for profile ",profileId)
	

	result, testcaseerr := Database.Query(fmt.Sprintf("SELECT Id,RelativePath,Executable from testcases WHERE CreatedBy='%s' AND ProfileId = %d",  username,profileId))
	if testcaseerr != nil {
	  	log.Println("Error Executing testcase SELECT ",testcaseerr.Error())
	  	return 
	}
	
	defer result.Close()	
	for result.Next() {
    var testcaserow Testcase  
    testcaseerr = result.Scan(&testcaserow.Id,&testcaserow.RelativePath,&testcaserow.Executable) 
    if testcaseerr != nil {
	  	log.Println("Error Executing testcase SELECT ",testcaseerr.Error())
	  	return 
	}
    log.Println("Id ",testcaserow.Id," Relative Path ",testcaserow.RelativePath," Executable ",testcaserow.Executable)
	}
	 
	
}

func  submit(rw http.ResponseWriter, r *http.Request) {
 
 
	log.Println("Calling execute command") 
	 
	var execrequest ExecRequest
	var qelement JobRequest
	
	decodeerr := json.NewDecoder(r.Body).Decode(&execrequest)
	if decodeerr != nil {
		http.Error(rw, decodeerr.Error(), http.StatusBadRequest)
		return
	}
	
	qelement.Profileid = execrequest.Profileid
	qelement.Username = execrequest.Username
	
	
	
	
	log.Println("Adding Profile ",qelement.Profileid," to execution list Requested by",qelement.Username)  
	jobQueue = append(jobQueue, qelement)	
	responseMessage := "submitted profile " + strconv.Itoa(qelement.Profileid) + "for execution by " + qelement.Username
	json.NewEncoder(rw).Encode(map[string]string{"response": responseMessage})
	
   //json.NewEncoder(rw).Encode(row)   
} 



func main() {
	
	
	currentTime := time.Now()
   f, err := os.OpenFile("/home/bala/goprojects/POC/EXECUTOR/log.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
   if err != nil {
	    log.Fatalf("error opening file: %v", err)
   }
   defer f.Close()
	
   log.SetOutput(f)
   log.Println("Starting the EXECUTOR REST API On SERVER ",currentTime)
    
    db, err := sql.Open("mysql", "E2EUSER:Automation$1234@tcp(192.168.1.151:3306)/E2E")
    if err != nil {
    	log.Println("Error opening SQL Connection",err.Error())
        panic(err.Error())
        return
    }    
    Database = db
    router := mux.NewRouter()    
    
    go PollQueue()
    
    router.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for preflight response
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	})
    
    //Route all the application related REST API 
    router.HandleFunc("/submit",submit).Methods("POST");  
    
	fmt.Println("Starting Server on port 9003......",currentTime)
    log.Fatal(http.ListenAndServe(":9003", router))  
    defer db.Close()
}