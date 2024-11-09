package main

import (
	"context"
	"fmt"
	"os/exec"
	"whattowatch/internal/config"
	"whattowatch/internal/loader"
	"whattowatch/internal/storage/postgresql"
	"whattowatch/pkg/logger"
)

func main() {
	printIP()
	// printData()

	cfg := config.MustLoad()

	log, file := logger.SetupLogger(cfg.Env, cfg.LogDir+"/loader")
	defer file.Close()

	postgresDB, err := postgresql.New(cfg, log)
	if err != nil {
		log.Error("creating storage error", "error", err.Error())
		panic("creating storage error: " + err.Error())
	}
	loader, err := loader.NewTMDbLoader(cfg, log, postgresDB)
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
