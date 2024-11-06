package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"whattowatch/internal/config"
	"whattowatch/internal/services/loader"
	"whattowatch/internal/storage"
	"whattowatch/pkg/logger"
)

func main() {
	printIP()
	// printData()

	cfg := config.MustLoad()

	log, file := logger.SetupLogger(cfg.Env, cfg.LogDir+"/loader")
	defer file.Close()

	storer, err := storage.New(cfg, log)
	if err != nil {
		log.Error("creating storage error", "error", err.Error())
		panic("creating storage error: " + err.Error())
	}
	loader, err := loader.NewTMDbLoader(cfg, log, storer)
	if err != nil {
		log.Error("creating loader error", "error", err.Error())
		panic("creating loader error: " + err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = loader.Load(ctx)
	if err != nil {
		log.Error("load error", "error", err.Error())
	}
}

func printIP() {
	app := "wget"

	arg0 := "-qO-"
	arg1 := "eth0.me"

	cmd := exec.Command(app, arg0, arg1)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Print the output
	fmt.Println("current IP is " + string(stdout))
}

func printData() {
	resp, err := http.Get("https://api.themoviedb.org/3/movie/11?api_key=12ea487afaad527386fc29d0b058cdbd")
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	// Print the output
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)
}
