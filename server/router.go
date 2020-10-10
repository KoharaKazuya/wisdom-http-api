package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func newRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.Timeout(50 * time.Second))

	r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Post("/comments", postComment)

	return r
}

func postComment(w http.ResponseWriter, r *http.Request) {
	// Read header
	cognitoUsername := r.Header.Get("wisdom-cognito-username")
	if cognitoUsername == "" {
		log.Print("[ERROR] lack of header wisdom-cognito-username")
		w.WriteHeader(500)
		return
	}

	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Printf("[ERROR] read request body: %v", err)
		w.WriteHeader(500)
		return
	}

	// Unmarshal
	var cReq CreateCommentRequest
	err = json.Unmarshal(b, &cReq)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`{"title":"json decode error"}`))
		return
	}

	// Validate request
	if cReq.PostID == "" {
		w.WriteHeader(400)
		w.Write([]byte(`{"title":"missing required field","detail":"postId is required"}`))
		return
	}
	if cReq.Content == "" {
		w.WriteHeader(400)
		w.Write([]byte(`{"title":"missing required field","detail":"content is required"}`))
		return
	}

	// Fetch user data
	cognitoUser, err := newCongnitoUserFromUsername(cognitoUsername)
	if err != nil {
		log.Printf("[ERROR] fetch user data: %v", err)
		w.WriteHeader(500)
		return
	}
	cReq.AuthorName = cognitoUser.name
	cReq.AuthorEmail = cognitoUser.email

	// construct env
	ssmSvc := ssm.New(awsSession)
	res, err := ssmSvc.GetParameterWithContext(r.Context(), &ssm.GetParameterInput{
		Name:           aws.String("/wisdom/wisdom-http-api/deploy-key"),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		log.Printf("[ERROR] fetch SSM parameter: %v", err)
		w.WriteHeader(500)
		return
	}
	env := Env{
		PrivateKeyPem: *res.Parameter.Value,
	}

	// inner proccess
	c, err := postCommentInner(r.Context(), cReq, env)
	if err != nil {
		log.Printf("[ERROR] post comment: %v", err)
		w.WriteHeader(500)
		return
	}

	// marshal
	output, err := json.Marshal(c)
	if err != nil {
		log.Printf("[ERROR] marshal response json: %v", err)
		w.WriteHeader(500)
		return
	}

	// response
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}
