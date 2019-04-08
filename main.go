package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/moguay/discordeq/applog"
	"github.com/moguay/discordeq/discord"
	"github.com/moguay/discordeq/listener"
	"github.com/xackery/eqemuconfig"
)

func main() {
	applog.StartupInteractive()
	log.SetOutput(applog.DefaultOutput)
	startService()
}

func startService() {
	log.Println("Starting DiscordEQ v0.51")
	var option string
	//Load config
	config, err := eqemuconfig.GetConfig()
	if err != nil {
		applog.Error.Println("Error while loading eqemu_config to start:", err.Error())
		fmt.Println("press a key then enter to exit.")

		fmt.Scan(&option)
		os.Exit(1)
	}
	if config.Discord.RefreshRate == 0 {
		config.Discord.RefreshRate = 10
	}

	if !isNewTelnetConfig(config) {
		log.Println("Telnet must be enabled for this tool to work. Check your eqemu_config, and please adjust.")
		fmt.Println("press a key then enter to exit.")
		fmt.Scan(&option)
		os.Exit(1)
	}

	if config.Discord.Username == "" {
		applog.Error.Println("I don't see a username set in your discord > username section of eqemu_config, please adjust.")
		fmt.Println("press a key then enter to exit.")
		fmt.Scan(&option)
		os.Exit(1)
	}

	if config.Discord.Password == "" && config.Discord.ClientID == "" {
		applog.Error.Println("I don't see a password set in your discord > password section of eqemu_config, as well as no client id, please adjust.")
		fmt.Println("press a key then enter to exit.")
		fmt.Scan(&option)
		os.Exit(1)
	}

	if config.Discord.ServerID == "" {
		applog.Error.Println("I don't see a serverid set in your discord > serverid section of eqemuconfig, please adjust.")
		fmt.Println("press a key then enter to exit.")
		fmt.Scan(&option)
		os.Exit(1)
	}

	if config.Discord.ChannelID == "" {
		applog.Error.Println("I don't see a channelid set in your  discord > channelid section of eqemuconfig.xml, please adjust.")
		fmt.Println("press a key then enter to exit.")
		fmt.Scan(&option)
		os.Exit(1)
	}
	disco := discord.Discord{}
	err = disco.Connect(config.Discord.Username, config.Discord.Password)
	if err != nil {
		applog.Error.Println("Error connecting to discord:", err.Error())
		fmt.Println("press a key then enter to exit.")
		fmt.Scan(&option)
		os.Exit(1)
	}
	go listenToDiscord(config, &disco)
	go listenToAUCTIONS(config, &disco)
	select {}
}

func isNewTelnetConfig(config *eqemuconfig.Config) bool {
	if strings.ToLower(config.World.Telnet.Enabled) == "true" {
		return true
	}
	if strings.ToLower(config.World.Tcp.Telnet) == "enabled" {
		return true
	}
	return false
}

func listenToDiscord(config *eqemuconfig.Config, disco *discord.Discord) (err error) {
	for {
		if len(config.Discord.Password) > 0 { //don't show username if it's token based
			applog.Info.Println("[Discord] Connecting as", config.Discord.Username, "...")
		} else {
			applog.Info.Println("[Discord] Connecting...")
		}
		if err = listener.ListenToDiscord(config, disco); err != nil {
			if strings.Contains(err.Error(), "Unauthorized") {
				applog.Info.Printf("Your bot is not authorized to access this server.\nClick this link and give the bot access: https://discordapp.com/oauth2/authorize?&client_id=%s&scope=bot&permissions=268446736", config.Discord.ClientID)
				return
			}
			applog.Error.Println("[Discord] Disconnected with error:", err.Error())
		}

		applog.Info.Println("[Discord] Reconnecting in 5 seconds...")
		time.Sleep(5 * time.Second)
		err = disco.Connect(config.Discord.Username, config.Discord.Password)
		if err != nil {
			applog.Error.Println("[Discord] Error connecting to discord:", err.Error())
		}
	}
}

func listenToAUCTIONS(config *eqemuconfig.Config, disco *discord.Discord) (err error) {
	for {
		listener.ListenToAUCTIONS(config, disco)
		applog.Info.Println("[AUCTIONS] Reconnecting in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}
