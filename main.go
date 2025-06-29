package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx := context.Background()
	sqlLogger := waLog.Stdout("SQL", "INFO", true)
	container, err := sqlstore.New(ctx, "sqlite3", "file:whatsapp.db?_foreign_keys=on", sqlLogger)
	if err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}

	client := whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "INFO", true))
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			// Jika menerima pesan masuk, auto-reply
			if v.Message.GetConversation() != "" {
				reply := fmt.Sprintf("Kamu berkata: %s", v.Message.GetConversation())
				msg := &waProto.Message{Conversation: &reply}
				_, _ = client.SendMessage(ctx, v.Info.Chat, msg)
			}
		}
	})

	if client.Store.ID == nil {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Masukkan nomor HP (format internasional, contoh: 6281234567890): ")
		phone, _ := reader.ReadString('\n')
		phone = strings.TrimSpace(phone)

		err = client.Connect()
		if err != nil {
			panic(err)
		}
		time.Sleep(2 * time.Second)

		showPushNotification := false
		clientType := whatsmeow.PairClientChrome
		clientDisplayName := "Chrome (Linux)"

		code, err := client.PairPhone(ctx, phone, showPushNotification, clientType, clientDisplayName)
		if err != nil {
			panic(err)
		}
		fmt.Println("=== Pairing Code WhatsApp ===")
		fmt.Printf("Ketik kode berikut di WhatsApp (Tautkan Perangkat > Masukkan Kode): %s\n", code)
		fmt.Println("=============================")
		fmt.Println("Tunggu proses pairing selesai...")
		select {}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		fmt.Println("Bot sudah online!")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	client.Disconnect()
}