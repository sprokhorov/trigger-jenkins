package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bndr/gojenkins"
)

var (
	jenkinsURL   string
	jenkinsUser  string
	jenkinsToken string
	jenkinsJob   string
)

func init() {
	jenkinsURL = lookupEnv("INPUT_JENKINS_URL")
	jenkinsUser = lookupEnv("INPUT_JENKINS_USER")
	jenkinsToken = lookupEnv("INPUT_JENKINS_TOKEN")
	jenkinsJob = lookupEnv("INPUT_JENKINS_JOB")
}

func lookupEnv(env string) string {
	value, ok := os.LookupEnv(env)
	if !ok {
		log.Fatalf("required variable %s was not defined", env)
	}
	return value
}

func lookupParams() map[string]string {
	result := map[string]string{}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "GITHUB") {
			pair := strings.Split(env, "=")
			result[pair[0]] = pair[1]
		}
	}
	return result
}

func main() {
	// create client
	ctx := context.Background()
	jenkins := gojenkins.CreateJenkins(nil, jenkinsURL, jenkinsUser, jenkinsToken)
	_, err := jenkins.Init(ctx)
	if err != nil {
		log.Fatalf("Failed init jenkins client, %+v", err)
	}
	// get job
	job, err := jenkins.GetJob(ctx, jenkinsJob)
	if err != nil {
		log.Fatalf("Failed to get Jenkins job '%s', %+v", jenkinsJob, err)
	}
	// invoke job
	log.Printf("Invoke Jenkins job '%s'", job.GetDetails().URL)
	queueid, err := job.InvokeSimple(ctx, lookupParams())
	if err != nil {
		log.Fatalf("Failed to invoke Jenkins job '%s', %+v", jenkinsJob, err)
	}

	log.Println("The new build was added to the queue")
	build, err := jenkins.GetBuildFromQueueID(ctx, queueid)
	if err != nil {
		log.Fatalf("Failed to get build of '%s' job, %+v", jenkinsJob, err)
	}

	log.Printf("Build '%s' is started", build.GetUrl())
	// watch the build status
	for build.IsRunning(ctx) {
		log.Println("The build is not finished yet. Waiting for 5 seconds.")
		time.Sleep(5000 * time.Millisecond)
		build.Poll(ctx)
	}
	log.Printf("Build number %d finished with result: %v", build.GetBuildNumber(), build.GetResult())
}
