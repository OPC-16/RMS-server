package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/OPC-16/RMS-server/model"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Signup(c echo.Context) error {
   //retrieve the redis client from the context
   rdb, ok := c.Get("redis").(*redis.Client)
   if !ok {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get Redis client from context"})
   }

   user := new (model.User)

   // Bind the request body to the user instance
   if err := c.Bind(user); err != nil {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Input"})
   }

   //perform basic validation
   if user.Name == "" || user.Email == "" || user.Password == "" {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name, Email and Password are required"})
   }

   // Serialize user to JSON
   userJson, err := json.Marshal(user)
   if err != nil {
      return fmt.Errorf("could not marshal user: %v", err)
   }

   // Extract the context.Context from echo.Context
   ctx := c.Request().Context()
   // Store JSON string in Redis with email as key
   err = rdb.Set(ctx, "user:" + user.Email, userJson, 0).Err()
   if err != nil {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to store user data in Redis"})
   }

   // Creating Sets for 'Admin's and 'Applicant's so that their retrieval would be easy
   userSetKey := fmt.Sprintf("userType:%s", user.UserType)
   err = rdb.SAdd(ctx, userSetKey, user.Email).Err()
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to store user data in Redis"})
   }

   //respond with success
   if user.UserType == "Applicant" {
      return c.JSON(http.StatusOK, map[string]string{"message": "Applicant signed up successfully"})
   } else {
      return c.JSON(http.StatusOK, map[string]string{"message": "Admin signed up successfully"})
   }
}

func Login(c echo.Context) error {
   type LoginRequest struct {
      Email    string `json:"email"`
      Password string `json:"password"`
   }

   req := new(LoginRequest)
   if err := c.Bind(req); err != nil {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Input"})
   }

   //Retrieve user details from redis
   ctx := c.Request().Context()
   rdb, ok := c.Get("redis").(*redis.Client)
   if !ok {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get Redis client from context"})
   }

   userJson, err := rdb.Get(ctx, "user:" + req.Email).Result()
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user details"})
   }

   //unmarshal user json
   var user model.User
   if err := json.Unmarshal([]byte(userJson), &user); err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal user data"})
   }

   //compare the passwords
   if req.Password != user.Password {
      return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid Credentials"})
   }

   //create JWT token
   token := jwt.New(jwt.SigningMethodHS256)
   claims := token.Claims.(jwt.MapClaims)
   claims["email"] = user.Email
   claims["usertype"] = user.UserType
   claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Token expires in 24 hours

   //sign the token with a secret
   tokenString, err := token.SignedString([]byte("secret"))
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate the token"})
   }

   // respond with JWT token
   return c.JSON(http.StatusOK, map[string]string{"token": tokenString})
}

func UploadResume(c echo.Context) error {
   apiEndPoint := "https://api.apilayer.com/resume_parser/upload"
   apiKey, _ := os.LookupEnv("API_KEY")
   fmt.Println(apiKey)
   filePath := "/home/omkar/resume.pdf"
   data, err := fetchAPIData(apiEndPoint, apiKey, filePath)
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
   }

   //TODO: don't return the data, store it in db in the model.Profile struct
   return c.JSON(http.StatusOK, data)
}

func PostJob(c echo.Context) error {
   rdb, ok := c.Get("redis").(*redis.Client)
   if !ok {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get Redis client from context"})
   }

   job := new(model.Job)

   if err := c.Bind(job); err != nil {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Input (Unable to Bind the request body to the job instance)"})
   }

   if job.Title == "" || job.Description == "" {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Job Title and Description are required"})
   }

   // setting the PostedOn field
   now := time.Now().UTC()
   job.PostedOn = &now

   // Retriving the user (admin) who posted this Job.
   // 1st we get the email of the user from the token claims, then as email is the key we used we retrieve the user from db and set the 'PostedBy' field
   email := c.Get("email").(string)
   ctx := c.Request().Context()
   userJson, err := rdb.Get(ctx, "user:" + email).Result()
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve the Admin details"})
   }

   var admin model.User
   if err := json.Unmarshal([]byte(userJson), &admin); err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal user data"})
   }

   //setting the PostedBy field
   job.PostedBy = admin

   jobJson, err := json.Marshal(job)
   if err != nil {
      return fmt.Errorf("could not marshal job: %v", err)
   }

   // Store JSON string in Redis with Title as key
   err = rdb.Set(ctx, "job:" + job.Title, jobJson, 0).Err()
   if err != nil {
      return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to store job data in Redis"})
   }

   // Creating Set for jobs so that their retrieval would be easy
   jobSetKey := "jobSet"
   err = rdb.SAdd(ctx, jobSetKey, job.Title).Err()
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to store job data in Redis"})
   }

   return c.JSON(http.StatusOK, map[string]string{"message": "Job Posted successfully"})
}

func fetchAPIData(apiEndPoint, apiKey, filePath string) (map[string]interface{}, error) {
   //open the file
   file, err := os.Open(filePath)
   if err != nil {
      return nil, err
   }
   defer file.Close()

   //create a buffer to hold the multipart form data
   body := &bytes.Buffer{}
   writer := multipart.NewWriter(body)

   //create a form field for the file
   part, err := writer.CreateFormFile("file", filepath.Base(filePath))
   if err != nil {
      return nil, err
   }

   //copy the file contents to the form field
   _, err = io.Copy(part, file)
   if err != nil {
      return nil, err
   }

   //close the writer to finalize the multipart form data
   err = writer.Close()
   if err != nil {
      return nil, err
   }

   //create a new request
   req, err := http.NewRequest("POST", apiEndPoint, body)
   if err != nil {
      return nil, err
   }

   //add the api key to the request header
   req.Header.Add("apikey", apiKey)
   req.Header.Add("Content-Type", "application/octet-stream")

   //make the http request
   client := &http.Client{}
   resp, err := client.Do(req)
   if err != nil {
      return nil, err
   }
   defer resp.Body.Close()

   //read the response body
   responseBody, err := io.ReadAll(resp.Body)
   if err != nil {
      return nil, err
   }

   //parse the json response
   var result map[string]interface{}
   err = json.Unmarshal(responseBody, &result)
   if err != nil {
      return nil, err
   }

   return result, nil
}

func ListApplicants(c echo.Context) error {
   rdb, ok := c.Get("redis").(*redis.Client)
   if !ok {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get Redis client from context"})
   }

   ctx := c.Request().Context()
   userSetKey := fmt.Sprintf("userType:%s", "Applicant")
   applicantEmailIDs, err := rdb.SMembers(ctx, userSetKey).Result()
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch Applicants"})
   }

   applicants := []model.User{}
   for _, appliapplicantEmailID := range applicantEmailIDs {
      userKey := fmt.Sprintf("user:%s", appliapplicantEmailID)
      userJson, err := rdb.Get(ctx, userKey).Result()
      if err != nil {
         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch Applicants"})
      }

      var user model.User
      err = json.Unmarshal([]byte(userJson), &user)
      if err != nil {
         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal user data"})
      }

      applicants = append(applicants, user)
   }

   return c.JSON(http.StatusOK, applicants)
}

func ListApplicant(c echo.Context) error {
   rdb, ok := c.Get("redis").(*redis.Client)
   if !ok {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get Redis client from context"})
   }
   ctx := c.Request().Context()

   applicantID := c.Param("applicant_id")
   result, err := rdb.Get(ctx, applicantID).Result()
   if err == redis.Nil {
      return c.JSON(http.StatusNotFound, map[string]string{"error": "Applicant not found"})
   } else if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch the applicant data"})
   }

   var applicant model.User
   err = json.Unmarshal([]byte(result), &applicant)
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse the applicant data"})
   }

   return c.JSON(http.StatusOK, applicant)
}

func ListJobs(c echo.Context) error {
   rdb, ok := c.Get("redis").(*redis.Client)
   if !ok {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get Redis client from context"})
   }
   ctx := c.Request().Context()

   jobSetKey := "jobSet"
   jobTitles, err := rdb.SMembers(ctx, jobSetKey).Result()
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch Jobs"})
   }

   jobs := []model.Job{}
   for _, jobTitle := range jobTitles {
      jobKey := fmt.Sprintf("job:%s", jobTitle)
      jobJson, err := rdb.Get(ctx, jobKey).Result()
      if err != nil {
         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch Jobs"})
      }

      var job model.Job
      err = json.Unmarshal([]byte(jobJson), &job)
      if err != nil {
         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unmarshal job data"})
      }

      jobs = append(jobs, job)
   }

   return c.JSON(http.StatusOK, jobs)
}

func ListJob(c echo.Context) error {
   rdb, ok := c.Get("redis").(*redis.Client)
   if !ok {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get Redis client from context"})
   }
   ctx := c.Request().Context()

   jobID := c.Param("job_id")
   result, err := rdb.Get(ctx, jobID).Result()
   if err == redis.Nil {
      return c.JSON(http.StatusNotFound, map[string]string{"error": "Job not found"})
   } else if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch the job data"})
   }

   var job model.Job
   err = json.Unmarshal([]byte(result), &job)
   if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse the job data"})
   }

   return c.JSON(http.StatusOK, job)
}
