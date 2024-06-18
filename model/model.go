package model

import "time"

type User struct {
   Name            string `json:"name"`
   Email           string `json:"email"`
   Address         string `json:"address"`
   UserType        string `json:"usertype"`
   Password        string `json:"password"`
   ProfileHeadline string `json:"profile_headline"`
}

type Profile struct {
   Applicant      User   `json:"applicant"`
   ResumeFileAddr string `json:"resume_file_address"`
   Skills         string `json:"skills"`
   Education      string `json:"education"`
   Experience     string `json:"experience"`
   Name           string `json:"name"`
   Email          string `json:"email"`
   Phone          string `json:"phone"`
}

type Job struct {
   Title             string     `json:"title"`
   Description       string     `json:"description"`
   PostedOn          *time.Time `json:"posted_on"`
   TotalApplications int        `json:"total_applications"`
   CompanyName       string     `json:"company_name"`
   PostedBy          User       `json:"posted_by"`
}
