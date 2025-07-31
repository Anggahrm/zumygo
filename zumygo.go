package main

import (
	"context"
	"fmt"
	"zumygo/handlers"
	"zumygo/helpers"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	_ "zumygo/commands/main"   // Import main commands
	_ "zumygo/commands/owner"  // Import owner commands
	_ "zumygo/commands/Auto"   // Import auto commands

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var log helpers.Logger

func init() {
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_EDGE.Enum()
	store.DeviceProps.Os = proto.String("Linux")
}

// connectWithRetry attempts to connect with retry mechanism
func connectWithRetry(conn *whatsmeow.Client, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		if err := conn.Connect(); err == nil {
			return nil
		}
		log.Warn(fmt.Sprintf("Connection attempt %d failed, retrying in %d seconds...", i+1, i+1))
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return fmt.Errorf("failed to connect after %d attempts", maxRetries)
}

func StartClient() {
	// Add recovery mechanism
	defer func() {
		if r := recover(); r != nil {
			log.Error(fmt.Sprintf("Recovered from panic: %v", r))
		}
	}()

	// Validate configuration
	if os.Getenv("OWNER") == "" {
		log.Error("OWNER environment variable is required")
		os.Exit(1)
	}
	if os.Getenv("PREFIX") == "" {
		log.Error("PREFIX environment variable is required")
		os.Exit(1)
	}

	log.Info("Configuration validation passed")
	log.Info("Starting WhatsApp bot...")
	
	ctx := context.Background()
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(ctx, "sqlite3", "file:session.db?_foreign_keys=on", dbLog)
	if err != nil {
		log.Error("Failed to create database container: " + err.Error())
		os.Exit(1)
	}

	handler := handlers.NewHandler(container)
	if handler == nil {
		log.Error("Failed to create message handler")
		os.Exit(1)
	}

	log.Info("Connecting Socket")
	conn := handler.Client()
	conn.PrePairCallback = func(jid types.JID, platform, businessName string) bool {
		log.Info("Connected Socket")
		return true
	}

	if conn.Store.ID == nil {
		// No ID stored, new login
		pairingNumber := os.Getenv("PAIRING_NUMBER")

		if pairingNumber != "" {
			pairingNumber = regexp.MustCompile(`\D+`).ReplaceAllString(pairingNumber, "")

			if err := connectWithRetry(conn, 3); err != nil {
				log.Error("Failed to connect for pairing: " + err.Error())
				os.Exit(1)
			}

			ctx := context.Background()
			code, err := conn.PairPhone(ctx, pairingNumber, true, whatsmeow.PairClientChrome, "Edge (Linux)")
			if err != nil {
				log.Error("Failed to pair phone: " + err.Error())
				os.Exit(1)
			}

			fmt.Println("Code Kamu : " + code)
		} else {
			qrChan, _ := conn.GetQRChannel(context.Background())
			if err := connectWithRetry(conn, 3); err != nil {
				log.Error("Failed to connect for QR: " + err.Error())
				os.Exit(1)
			}

			for evt := range qrChan {
				switch string(evt.Event) {
				case "code":
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
					log.Info("Qr Required")
				}
			}
		}
	} else {
		// Already logged in, just connect
		if err := connectWithRetry(conn, 3); err != nil {
			log.Error("Failed to connect: " + err.Error())
			os.Exit(1)
		}
		log.Info("Connected Socket")
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Info("Shutting down gracefully...")
	conn.Disconnect()
}
