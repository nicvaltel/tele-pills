package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/ssh"
)

var envFile string = "deploy.env"

func connectSSH() *ssh.Client {

	errEnv := godotenv.Load(envFile)
	if errEnv != nil {
		log.Panic(fmt.Sprintf("Error loading %s file.\n", envFile))
	}

	host := os.Getenv("DEPLOY_HOST")
	port := os.Getenv("DEPLOY_PORT")
	user := os.Getenv("DEPLOY_USER")
	keyPath := os.Getenv("DEPLOY_KEYPATH")

	privateKey, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	conn, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	return conn
}

func execSSHCommands(conn *ssh.Client, commands []string) {

	var buff bytes.Buffer

	for _, cmd := range commands {
		session, err := conn.NewSession()
		if err != nil {
			log.Fatalf("Failed to create session: %v", err)
		}
		defer session.Close()

		session.Stdout = &buff

		if err := session.Run(cmd); err != nil {
			log.Fatalf("Failed to run command '%s': %v", cmd, err)
		}
		fmt.Printf("Command '%s' executed successfully\n", cmd)
		fmt.Println(buff.String())
	}

}

func copyFileSSH(sshClient *ssh.Client, localFile string, remoteFilePath string, permissions string) {

	client, err := scp.NewClientBySSH(sshClient)
	if err != nil {
		fmt.Println("Error creating new SSH session from existing connection", err)
	}

	f, _ := os.Open(localFile)

	defer client.Close()

	defer f.Close()

	err = client.CopyFile(context.Background(), f, remoteFilePath, permissions)

	if err != nil {
		fmt.Println("Error while copying file ", err)
	}
}

func RunSSH() {

	// cmd := exec.Command("rm ../../Pills && go build ../..")
	// err := cmd.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	conn := connectSSH()
	defer conn.Close()

	commands := []string{
		"mkdir -p /home/kolay/bot/Pills", // -p flag create if not exist
		"mkdir -p /home/kolay/bot/Pills/database/postgresql/migration",
		"touch -a /home/kolay/bot/Pills/database/postgresql/migration/migration.md5", // -a flag for don't owerwrite if exist
	}
	execSSHCommands(conn, commands)

	copyFileSSH(conn, "../../Pills", "/home/kolay/bot/Pills/Pills-exe", "0744")
	copyFileSSH(conn, "../../Pills", "/home/kolay/bot/Pills/Pills-exe", "0744")
	copyFileSSH(conn, "../../config.env", "/home/kolay/bot/Pills/config.env", "0644")
	copyFileSSH(conn, "../../database/postgresql/migration/migration.sql", "/home/kolay/bot/Pills/database/postgresql/migration/migration.sql", "0644")

	commands = []string{
		"touch /home/kolay/bot/Pills/run.sh",
		"chmod +x /home/kolay/bot/Pills/run.sh",
		"echo -e '#!/bin/bash \n\ncd /home/kolay/bot/Pills/ \n./Pills-exe' > /home/kolay/bot/Pills/run.sh", // -e flag for new lines
		"touch /etc/systemd/system/pillsbot.service",
		"echo -e '[Unit]\nDescription=Telegram bot pills reminder\nAfter=default.target\nStartLimitIntervalSec=0\n\n[Service]\nRestart=always\nRestartSec=3\nExecStart=/home/kolay/bot/Pills/run.sh\n\n[Install]\nWantedBy=default.target' > /etc/systemd/system/pillsbot.service",
		"systemctl daemon-reload",
		"systemctl enable pillsbot.service",
		"systemctl start pillsbot.service",
	}
	execSSHCommands(conn, commands)

}

func main() {
	RunSSH()
}
