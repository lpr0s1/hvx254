package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas" // Import ajouté pour le canvas.Text
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type BotFeatures struct {
	Ping               bool
	Help               bool
	News               bool
	Faq                bool
	Welcome            bool
	Goodbye            bool
	AutoReply          bool
	Jokes              bool
	Quotes             bool
	SimpleModeration   bool
	AntiSpam           bool
	Logs               bool
	MessageCounter     bool
	MemberCounter      bool
	AutoReaction       bool
	PrefixCommands     bool
	SimpleSlash        bool
	DmWelcome          bool
	DmInfo             bool
	Reminders          bool
	Polls              bool
	AutoRoles          bool
	ReactionRoles      bool
	SimpleMusic        bool
	ServerInfo         bool
	UserInfo           bool
	Uptime             bool
	DayStats           bool
	WeekStats          bool
	MonthStats         bool
	Announce           bool
}

type CustomCommand struct {
	Name    string
	Trigger string
	Reply   string
}

type BotConfig struct {
	Token       string
	Name        string
	Prefix      string
	StatusText  string
	StatusType  string
	Description string
	Features    BotFeatures
	Commands    []CustomCommand
}

type BotState struct {
	Session      *discordgo.Session
	Running      bool
	StartTime    time.Time
	MessageCount int64
	MemberCount  int64
	Mutex        sync.Mutex
}

var (
	cfg = BotConfig{
		Prefix:     "!",
		StatusText: "Online",
		StatusType: "PLAYING",
	}
	state = BotState{}
)

func safeRun(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %v\n", r)
			debug.PrintStack()
		}
	}()
	fn()
}

func startBot(logFn func(string), statusFn func(bool)) {
	safeRun(func() {
		state.Mutex.Lock()
		if state.Running {
			state.Mutex.Unlock()
			return
		}
		token := strings.TrimSpace(cfg.Token)
		if token == "" {
			state.Mutex.Unlock()
			logFn("Token manquant")
			return
		}
		s, err := discordgo.New("Bot " + token)
		if err != nil {
			state.Mutex.Unlock()
			logFn("Erreur creation session")
			return
		}
		s.Identify.Intents = discordgo.IntentsGuilds |
			discordgo.IntentsGuildMessages |
			discordgo.IntentsGuildMembers |
			discordgo.IntentsMessageContent

		s.AddHandler(func(ses *discordgo.Session, r *discordgo.Ready) {
			logFn("Bot connecte en tant que " + ses.State.User.Username)
			state.Mutex.Lock()
			state.StartTime = time.Now()
			state.Running = true
			state.Mutex.Unlock()
			statusFn(true)
			setPresence(ses, logFn)
		})

		s.AddHandler(func(ses *discordgo.Session, m *discordgo.MessageCreate) {
			handleMessage(ses, m, logFn)
		})

		s.AddHandler(func(ses *discordgo.Session, m *discordgo.GuildMemberAdd) {
			handleMemberJoin(ses, m, logFn)
		})

		s.AddHandler(func(ses *discordgo.Session, m *discordgo.GuildMemberRemove) {
			handleMemberLeave(ses, m, logFn)
		})

		err = s.Open()
		if err != nil {
			state.Mutex.Unlock()
			logFn("Erreur connexion bot")
			return
		}

		state.Session = s
		state.Running = true
		state.StartTime = time.Now()
		state.Mutex.Unlock()
		logFn("Bot demarre")
		statusFn(true)
	})
}

func stopBot(logFn func(string), statusFn func(bool)) {
	safeRun(func() {
		state.Mutex.Lock()
		if !state.Running || state.Session == nil {
			state.Mutex.Unlock()
			logFn("Bot deja arrete")
			return
		}
		err := state.Session.Close()
		if err != nil {
			logFn("Erreur arret bot")
		}
		state.Session = nil
		state.Running = false
		state.Mutex.Unlock()
		logFn("Bot arrete")
		statusFn(false)
	})
}

func setPresence(s *discordgo.Session, logFn func(string)) {
	t := discordgo.ActivityTypeGame // CORRECTION : ActivityTypeGame au lieu de ActivityTypePlaying
	switch cfg.StatusType {
	case "WATCHING":
		t = discordgo.ActivityTypeWatching
	case "LISTENING":
		t = discordgo.ActivityTypeListening
	case "COMPETING":
		t = discordgo.ActivityTypeCompeting
	}
	err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: cfg.StatusText,
				Type: t,
			},
		},
		Status: "online",
	})
	if err != nil {
		logFn("Erreur statut")
	}
}

func handleMessage(s *discordgo.Session, m *discordgo.MessageCreate, logFn func(string)) {
	if m.Author.Bot {
		return
	}

	state.Mutex.Lock()
	state.MessageCount++
	state.Mutex.Unlock()

	if cfg.Features.Logs {
		logFn("Message de " + m.Author.Username + ": " + m.Content)
	}

	if cfg.Features.SimpleModeration {
		bad := []string{"insulte1", "insulte2"}
		lower := strings.ToLower(m.Content)
		for _, w := range bad {
			if strings.Contains(lower, w) {
				_ = s.ChannelMessageDelete(m.ChannelID, m.ID)
				_, _ = s.ChannelMessageSend(m.ChannelID, "Message supprime")
				return
			}
		}
	}

	if cfg.Features.AutoReaction {
		if strings.Contains(strings.ToLower(m.Content), "bonjour") {
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, "👋")
		}
	}

	if cfg.Features.AutoReply {
		if strings.Contains(strings.ToLower(m.Content), "comment ca va") {
			_, _ = s.ChannelMessageSendReply(m.ChannelID, "Je vais bien", m.Reference())
		}
	}

	if cfg.Features.Jokes && strings.Contains(strings.ToLower(m.Content), "blague") {
		_, _ = s.ChannelMessageSendReply(m.ChannelID, "Blague simple", m.Reference())
	}

	if cfg.Features.Quotes && strings.Contains(strings.ToLower(m.Content), "citation") {
		_, _ = s.ChannelMessageSendReply(m.ChannelID, "Citation simple", m.Reference())
	}

	if cfg.Features.PrefixCommands && strings.HasPrefix(m.Content, cfg.Prefix) {
		content := strings.TrimPrefix(m.Content, cfg.Prefix)
		parts := strings.Fields(content)
		if len(parts) == 0 {
			return
		}
		cmd := strings.ToLower(parts[0])

		if cfg.Features.Ping && cmd == "ping" {
			_, _ = s.ChannelMessageSendReply(m.ChannelID, "Pong", m.Reference())
		}

		if cfg.Features.Help && cmd == "help" {
			_, _ = s.ChannelMessageSendReply(m.ChannelID, "Commandes: "+listCommands(), m.Reference())
		}

		if cfg.Features.News && cmd == "news" {
			_, _ = s.ChannelMessageSendReply(m.ChannelID, "Pas de news", m.Reference())
		}

		if cfg.Features.Faq && cmd == "faq" {
			_, _ = s.ChannelMessageSendReply(m.ChannelID, "Pas de faq", m.Reference())
		}

		if cfg.Features.ServerInfo && cmd == "server" {
			if m.GuildID != "" {
				g, err := s.State.Guild(m.GuildID)
				if err == nil {
					_, _ = s.ChannelMessageSendReply(m.ChannelID, "Serveur: "+g.Name, m.Reference())
				}
			}
		}

		if cfg.Features.UserInfo && cmd == "me" {
			_, _ = s.ChannelMessageSendReply(m.ChannelID, "Tu es "+m.Author.Username, m.Reference())
		}

		for _, c := range cfg.Commands {
			if m.Content == c.Trigger {
				_, _ = s.ChannelMessageSendReply(m.ChannelID, c.Reply, m.Reference())
			}
		}
	}
}

func listCommands() string {
	var list []string
	if cfg.Features.Ping {
		list = append(list, cfg.Prefix+"ping")
	}
	if cfg.Features.Help {
		list = append(list, cfg.Prefix+"help")
	}
	if cfg.Features.News {
		list = append(list, cfg.Prefix+"news")
	}
	if cfg.Features.Faq {
		list = append(list, cfg.Prefix+"faq")
	}
	if cfg.Features.ServerInfo {
		list = append(list, cfg.Prefix+"server")
	}
	if cfg.Features.UserInfo {
		list = append(list, cfg.Prefix+"me")
	}
	for _, c := range cfg.Commands {
		list = append(list, c.Trigger)
	}
	return strings.Join(list, ", ")
}

func handleMemberJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd, logFn func(string)) {
	if cfg.Features.Welcome {
		chID := findDefaultChannel(s, m.GuildID)
		if chID != "" {
			_, _ = s.ChannelMessageSend(chID, "Bienvenue "+m.User.Username)
		}
	}
	if cfg.Features.DmWelcome {
		// CORRECTION : Création du canal privé (DM) avant l'envoi du message
		dmChannel, err := s.UserChannelCreate(m.User.ID)
		if err == nil {
			_, _ = s.ChannelMessageSend(dmChannel.ID, "Bienvenue sur le serveur")
		}
	}
}

func handleMemberLeave(s *discordgo.Session, m *discordgo.GuildMemberRemove, logFn func(string)) {
	if cfg.Features.Goodbye {
		chID := findDefaultChannel(s, m.GuildID)
		if chID != "" {
			_, _ = s.ChannelMessageSend(chID, m.User.Username+" a quitte le serveur")
		}
	}
}

func findDefaultChannel(s *discordgo.Session, guildID string) string {
	g, err := s.State.Guild(guildID)
	if err != nil {
		return ""
	}
	if g.SystemChannelID != "" {
		return g.SystemChannelID
	}
	for _, ch := range g.Channels {
		if ch.Type == discordgo.ChannelTypeGuildText {
			return ch.ID
		}
	}
	return ""
}

func checkToken(token string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return false
	}
	s.Identify.Intents = discordgo.IntentsGuilds
	err = s.Open()
	if err != nil {
		return false
	}
	_ = s.Close()
	return true
}

func main() {
	log.SetOutput(os.Stdout)

	a := app.NewWithID("bot.discord.creator")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Createur de bot Discord")
	w.Resize(fyne.NewSize(1100, 700))

	logBuffer := ""
	logMutex := sync.Mutex{}

	appendLog := func(s string) {
		logMutex.Lock()
		defer logMutex.Unlock()
		t := time.Now().Format("15:04:05")
		line := fmt.Sprintf("[%s] %s", t, s)
		if logBuffer == "" {
			logBuffer = line
		} else {
			logBuffer += "\n" + line
		}
	}

	logsWidget := widget.NewMultiLineEntry()
	logsWidget.SetPlaceHolder("En attente de logs...")
	logsWidget.Disable()

	statusLabel := widget.NewLabel("Hors ligne")

	// CORRECTION : Utilisation de canvas.NewText à la place de widget.NewLabel
	statusDot := canvas.NewText("●", theme.ErrorColor())
	statusDot.TextStyle.Bold = true
	statusDot.TextStyle.Monospace = true
	statusDot.TextSize = 14

	setOnline := func(on bool) {
		if on {
			statusLabel.SetText("En ligne")
			statusDot.Color = theme.PrimaryColor() // Modifié
		} else {
			statusLabel.SetText("Hors ligne")
			statusDot.Color = theme.ErrorColor() // Modifié
		}
		statusDot.Refresh() // Essentiel pour mettre à jour l'affichage dans Fyne
	}

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("Colle ici le token du bot")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Nom du bot")

	prefixEntry := widget.NewEntry()
	prefixEntry.SetText("!")

	statusTextEntry := widget.NewEntry()
	statusTextEntry.SetText("Online")

	statusTypeSelect := widget.NewSelect([]string{"PLAYING", "WATCHING", "LISTENING", "COMPETING"}, func(string) {})
	statusTypeSelect.SetSelected("PLAYING")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Decris a quoi sert ton bot")

	featureCheck := func(label string, get func() bool, set func(bool)) *widget.Check {
		c := widget.NewCheck(label, func(b bool) {
			set(b)
		})
		c.SetChecked(get())
		return c
	}

	cfg.Features.Ping = true
	cfg.Features.Help = true
	cfg.Features.Logs = true
	cfg.Features.PrefixCommands = true

	featuresLeft := container.NewVBox(
		featureCheck("Repondre a ping", func() bool { return cfg.Features.Ping }, func(b bool) { cfg.Features.Ping = b }),
		featureCheck("Commande help", func() bool { return cfg.Features.Help }, func(b bool) { cfg.Features.Help = b }),
		featureCheck("Commande news", func() bool { return cfg.Features.News }, func(b bool) { cfg.Features.News = b }),
		featureCheck("Commande faq", func() bool { return cfg.Features.Faq }, func(b bool) { cfg.Features.Faq = b }),
		featureCheck("Message bienvenue", func() bool { return cfg.Features.Welcome }, func(b bool) { cfg.Features.Welcome = b }),
		featureCheck("Message au revoir", func() bool { return cfg.Features.Goodbye }, func(b bool) { cfg.Features.Goodbye = b }),
		featureCheck("Reponses auto simples", func() bool { return cfg.Features.AutoReply }, func(b bool) { cfg.Features.AutoReply = b }),
		featureCheck("Blagues", func() bool { return cfg.Features.Jokes }, func(b bool) { cfg.Features.Jokes = b }),
		featureCheck("Citations", func() bool { return cfg.Features.Quotes }, func(b bool) { cfg.Features.Quotes = b }),
		featureCheck("Moderation simple", func() bool { return cfg.Features.SimpleModeration }, func(b bool) { cfg.Features.SimpleModeration = b }),
		featureCheck("Anti spam basique", func() bool { return cfg.Features.AntiSpam }, func(b bool) { cfg.Features.AntiSpam = b }),
		featureCheck("Logs du bot", func() bool { return cfg.Features.Logs }, func(b bool) { cfg.Features.Logs = b }),
		featureCheck("Compteur messages", func() bool { return cfg.Features.MessageCounter }, func(b bool) { cfg.Features.MessageCounter = b }),
		featureCheck("Compteur membres", func() bool { return cfg.Features.MemberCounter }, func(b bool) { cfg.Features.MemberCounter = b }),
		featureCheck("Reactions auto", func() bool { return cfg.Features.AutoReaction }, func(b bool) { cfg.Features.AutoReaction = b }),
	)

	featuresRight := container.NewVBox(
		featureCheck("Commandes prefix", func() bool { return cfg.Features.PrefixCommands }, func(b bool) { cfg.Features.PrefixCommands = b }),
		featureCheck("Commandes slash simples", func() bool { return cfg.Features.SimpleSlash }, func(b bool) { cfg.Features.SimpleSlash = b }),
		featureCheck("DM bienvenue", func() bool { return cfg.Features.DmWelcome }, func(b bool) { cfg.Features.DmWelcome = b }),
		featureCheck("DM info", func() bool { return cfg.Features.DmInfo }, func(b bool) { cfg.Features.DmInfo = b }),
		featureCheck("Rappels simples", func() bool { return cfg.Features.Reminders }, func(b bool) { cfg.Features.Reminders = b }),
		featureCheck("Sondages", func() bool { return cfg.Features.Polls }, func(b bool) { cfg.Features.Polls = b }),
		featureCheck("Roles auto", func() bool { return cfg.Features.AutoRoles }, func(b bool) { cfg.Features.AutoRoles = b }),
		featureCheck("Roles par reaction", func() bool { return cfg.Features.ReactionRoles }, func(b bool) { cfg.Features.ReactionRoles = b }),
		featureCheck("Musique simple", func() bool { return cfg.Features.SimpleMusic }, func(b bool) { cfg.Features.SimpleMusic = b }),
		featureCheck("Infos serveur", func() bool { return cfg.Features.ServerInfo }, func(b bool) { cfg.Features.ServerInfo = b }),
		featureCheck("Infos utilisateur", func() bool { return cfg.Features.UserInfo }, func(b bool) { cfg.Features.UserInfo = b }),
		featureCheck("Temps en ligne", func() bool { return cfg.Features.Uptime }, func(b bool) { cfg.Features.Uptime = b }),
		featureCheck("Stats jour", func() bool { return cfg.Features.DayStats }, func(b bool) { cfg.Features.DayStats = b }),
		featureCheck("Stats semaine", func() bool { return cfg.Features.WeekStats }, func(b bool) { cfg.Features.WeekStats = b }),
		featureCheck("Stats mois", func() bool { return cfg.Features.MonthStats }, func(b bool) { cfg.Features.MonthStats = b }),
		featureCheck("Messages annonce", func() bool { return cfg.Features.Announce }, func(b bool) { cfg.Features.Announce = b }),
	)

	cmdNameEntry := widget.NewEntry()
	cmdNameEntry.SetPlaceHolder("Nom de la commande")

	cmdTriggerEntry := widget.NewEntry()
	cmdTriggerEntry.SetPlaceHolder("Texte a taper")

	cmdReplyEntry := widget.NewMultiLineEntry()
	cmdReplyEntry.SetPlaceHolder("Reponse du bot")

	commandsList := widget.NewMultiLineEntry()
	commandsList.Disable()
	commandsList.SetPlaceHolder("Aucune commande pour le moment")

	refreshCommandsView := func() {
		if len(cfg.Commands) == 0 {
			commandsList.SetText("Aucune commande pour le moment")
			return
		}
		var lines []string
		for i, c := range cfg.Commands {
			lines = append(lines, fmt.Sprintf("%d. %s (%s) -> %s", i+1, c.Name, c.Trigger, c.Reply))
		}
		commandsList.SetText(strings.Join(lines, "\n"))
	}

	addCmdBtn := widget.NewButton("Ajouter la commande", func() {
		name := strings.TrimSpace(cmdNameEntry.Text)
		trig := strings.TrimSpace(cmdTriggerEntry.Text)
		rep := strings.TrimSpace(cmdReplyEntry.Text)
		if name == "" || trig == "" || rep == "" {
			return
		}
		cfg.Commands = append(cfg.Commands, CustomCommand{
			Name:    name,
			Trigger: trig,
			Reply:   rep,
		})
		cmdNameEntry.SetText("")
		cmdTriggerEntry.SetText("")
		cmdReplyEntry.SetText("")
		refreshCommandsView()
	})

	tokenStatusLabel := widget.NewLabel("En attente de verification")
	checkTokenBtn := widget.NewButton("Verifier le token", func() {
		t := strings.TrimSpace(tokenEntry.Text)
		if t == "" {
			tokenStatusLabel.SetText("Token vide")
			return
		}
		tokenStatusLabel.SetText("Verification en cours...")
		go func() {
			ok := checkToken(t)
			if ok {
				tokenStatusLabel.SetText("Token valide")
			} else {
				tokenStatusLabel.SetText("Token invalide")
			}
		}()
	})

	startBtn := widget.NewButton("Demarrer le bot", func() {
		cfg.Token = tokenEntry.Text
		cfg.Name = nameEntry.Text
		if strings.TrimSpace(prefixEntry.Text) == "" {
			cfg.Prefix = "!"
		} else {
			cfg.Prefix = prefixEntry.Text
		}
		if strings.TrimSpace(statusTextEntry.Text) == "" {
			cfg.StatusText = "Online"
		} else {
			cfg.StatusText = statusTextEntry.Text
		}
		cfg.StatusType = statusTypeSelect.Selected
		cfg.Description = descEntry.Text

		go startBot(func(s string) {
			appendLog(s)
			logsWidget.SetText(logBuffer)
			logsWidget.CursorRow = len(strings.Split(logBuffer, "\n")) - 1
		}, func(on bool) {
			setOnline(on)
		})
	})

	stopBtn := widget.NewButton("Arreter le bot", func() {
		go stopBot(func(s string) {
			appendLog(s)
			logsWidget.SetText(logBuffer)
		}, func(on bool) {
			setOnline(on)
		})
	})

	refreshLogsBtn := widget.NewButton("Actualiser les logs", func() {
		logsWidget.SetText(logBuffer)
	})

	topBar := container.NewHBox(
		statusDot,
		widget.NewLabel("Createur de bot Discord"),
		layout.NewSpacer(),
		statusLabel,
	)

	tokenRow := container.NewHBox(
		tokenEntry,
		checkTokenBtn,
	)

	infoForm := container.NewVBox(
		widget.NewLabel("Connexion et infos"),
		tokenRow,
		tokenStatusLabel,
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			container.NewVBox(
				widget.NewLabel("Nom du bot"),
				nameEntry,
			),
			container.NewVBox(
				widget.NewLabel("Prefixe"),
				prefixEntry,
			),
		),
		container.NewGridWithColumns(2,
			container.NewVBox(
				widget.NewLabel("Texte de statut"),
				statusTextEntry,
			),
			container.NewVBox(
				widget.NewLabel("Type de statut"),
				statusTypeSelect,
			),
		),
		widget.NewLabel("Description"),
		descEntry,
	)

	featuresBox := container.NewVBox(
		widget.NewLabel("Options du bot"),
		container.NewGridWithColumns(2, featuresLeft, featuresRight),
	)

	commandsBox := container.NewVBox(
		widget.NewLabel("Commandes personnalisees"),
		widget.NewLabel("Nom"),
		cmdNameEntry,
		widget.NewLabel("Texte a taper"),
		cmdTriggerEntry,
		widget.NewLabel("Reponse"),
		cmdReplyEntry,
		addCmdBtn,
		widget.NewLabel("Liste des commandes"),
		commandsList,
	)

	controlBox := container.NewVBox(
		widget.NewLabel("Logs et controle"),
		logsWidget,
		container.NewHBox(
			refreshLogsBtn,
			layout.NewSpacer(),
			stopBtn,
			startBtn,
		),
	)

	leftCol := container.NewVBox(
		infoForm,
		widget.NewSeparator(),
		featuresBox,
	)

	rightCol := container.NewVBox(
		commandsBox,
		widget.NewSeparator(),
		controlBox,
	)

	mainSplit := container.NewHSplit(leftCol, rightCol)
	mainSplit.Offset = 0.55

	root := container.NewBorder(topBar, nil, nil, nil, mainSplit)
	w.SetContent(root)

	w.SetCloseIntercept(func() {
		if state.Running {
			stopBot(func(s string) {}, func(bool) {})
		}
		w.Close()
		a.Quit()
	})

	w.ShowAndRun()
}
