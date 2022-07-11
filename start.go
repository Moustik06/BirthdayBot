package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mattn/go-sqlite3"
)

var (
	db, _  = sql.Open("sqlite3", "birthday.db")
)

func prepareExec(username string, Date string) {
	stmt, err := db.Prepare("INSERT INTO birthday(Username,Date) VALUES(?, ?)")
	checkError(err)
	defer stmt.Close()
	stmt.Exec(username, Date)
}
func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}
func ConnectToDiscord() {
	sqlite3.Version()

	discord, err := discordgo.New("Bot " + "")
	if err != nil {
		log.Panic("Erreur pendant la création de session")
		return
	}
	discord.AddHandler(messageCreate)
	discord.Identify.Intents = discordgo.IntentGuildMessages

	err = discord.Open()
	if err != nil {
		log.Panic("Erreur de connexion")
	}
	log.Println("Lancement du bot")
	discord.UpdateGameStatus(0, "Tempest birthday bot")
	checkBirthday(discord)
	defer discord.Close()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

	command := strings.Split(message.Content, " ")[0]

	log.Println("Commande : " + command)
	log.Print("args : ")

	switch command {
	case "birthday-add":
		args := strings.Split(message.Content, " ")[1:]
		if len(args) > 1 {
			log.Panic("Too many args")
			return
		}
		prepareExec(message.Author.ID, args[0])
		session.ChannelMessageSend(message.ChannelID, "Your birthday has been added")

	case "birthday-next":
		rows, err := db.Query("SELECT * FROM birthday WHERE strftime('%m-%d',birthday.Date ) > strftime('%m-%d',date('now', 'localtime')) ORDER BY birthday.Date ASC LIMIT 1")
		checkError(err)
		defer rows.Close()
		for rows.Next() {
			var (
				id   string
				date string
			)
			if err := rows.Scan(&id, &date); err != nil {
				log.Fatal(err)
			}

			User, err := session.User(id)
			checkError(err)

			log.Printf("The closest birthday is %s the %s", User, date)
			session.ChannelMessageSend(message.ChannelID, "The closest birthday is "+User.Username+" the "+date)
		}
	}

}

func checkBirthday(session *discordgo.Session) {
	rows, err := db.Query("SELECT * FROM birthday WHERE strftime('%m-%d',birthday.Date ) = strftime('%m-%d',date('now', 'localtime'))")
	checkError(err)
	defer rows.Close()
	for rows.Next() {
		var (
			id   string
			date string
		)
		if err := rows.Scan(&id, &date); err != nil {
			log.Fatal(err)
		}
		chanID, err := session.Channel("996041744653221930")
		checkError(err)
		session.ChannelMessageSend(chanID.ID, "Today it's the birthday of <@"+id+"> !!!")
	}

	temps := time.Until(time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour))
	time.AfterFunc(temps, func() { checkBirthday(session) })
}
