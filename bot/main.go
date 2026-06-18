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
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type BotFeatures struct {
	Ping             bool
	Help             bool
	News             bool
	Faq              bool
	Welcome          bool
	Goodbye          bool
	AutoReply        bool
	Jokes            bool
	Quotes           bool
	SimpleModeration bool
	AntiSpam         bool
	Logs             bool
	MessageCounter   bool
	MemberCounter    bool
	AutoReaction     bool
	PrefixCommands   bool
	SimpleSlash      bool
	DmWelcome        bool
	DmInfo           bool
	Reminders        bool
	Polls            bool
	AutoRoles        bool
	ReactionRoles    bool
	SimpleMusic      bool
	ServerInfo       bool
	UserInfo         bool
	Uptime           bool
	DayStats         bool
	WeekStats        bool
	MonthStats       bool
	Announce         bool
}

type CustomCommand struct {
	Name    string
	Trigger string
	Reply   string
}

type BotConfig struct {
	Token         string
	Name          string
	Prefix        string
	StatusText    string
	StatusType    string
	Description   string
	TargetUserTag string
	RSSFeedURL    string
	Features      BotFeatures
	Commands      []CustomCommand
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
		StatusText: "En ligne",
		StatusType: "PLAYING",
	}
	state = BotState{}
)

func safeRun(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Erreur critique: %v\n", r)
			debug.PrintStack()
		}
	}()
	fn()
}

func startBot(logFn func(string), statusFn func(bool), targetStatusFn func(string)) {
	safeRun(func() {
		state.Mutex.Lock()
		if state.Running {
			state.Mutex.Unlock()
			return
		}
		token := strings.TrimSpace(cfg.Token)
		if token == "" {
			state.Mutex.Unlock()
			logFn("Erreur : Token manquant.")
			return
		}
		s, err := discordgo.New("Bot " + token)
		if err != nil {
			state.Mutex.Unlock()
			logFn("Erreur lors de la creation de la session.")
			return
		}
		
		s.Identify.Intents = discordgo.IntentsGuilds |
			discordgo.IntentsGuildMessages |
			discordgo.IntentsGuildMembers |
			discordgo.IntentsMessageContent

		s.AddHandler(func(ses *discordgo.Session, r *discordgo.Ready) {
			logFn("Bot connecte avec succes en tant que : " + ses.State.User.Username)
			state.Mutex.Lock()
			state.StartTime = time.Now()
			state.Running = true
			state.Mutex.Unlock()
			statusFn(true)
			setPresence(ses, logFn)

			go checkAdminPerms(ses, logFn)
			go checkTargetUser(ses, cfg.TargetUserTag, targetStatusFn, logFn)
		})

		s.AddHandler(func(ses *discordgo.Session, m *discordgo.MessageCreate) {
			handleMessage(ses, m, logFn)
		})

		s.AddHandler(func(ses *discordgo.Session, h *discordgo.GuildMemberAdd) {
			handleNewMember(ses, h, logFn)
		})

		err = s.Open()
		if err != nil {
			state.Mutex.Unlock()
			logFn("Impossible de connecter le bot. Verifie tes Privileged Intents sur le portail developpeur.")
			return
		}

		state.Session = s
		state.Running = true
		state.StartTime = time.Now()
		state.Mutex.Unlock()
		logFn("Lancement du processus principal reussi.")
		statusFn(true)
	})
}

func checkAdminPerms(s *discordgo.Session, logFn func(string)) {
	time.Sleep(3 * time.Second)
	hasAdmin := false

	for _, g := range s.State.Guilds {
		member, err := s.State.Member(g.ID, s.State.User.ID)
		if err != nil {
			continue
		}
		for _, roleID := range member.Roles {
			role, err := s.State.Role(g.ID, roleID)
			if err == nil && (role.Permissions&discordgo.PermissionAdministrator) != 0 {
				hasAdmin = true
				break
			}
		}
		if hasAdmin {
			break
		}
	}

	if hasAdmin {
		logFn("Info: Permission ADMINISTRATEUR detectee.")
	} else {
		logFn("Info: Le bot n'a pas la permission Administrateur sur les serveurs actuels.")
	}
}

func checkTargetUser(s *discordgo.Session, targetTag string, statusFn func(string), logFn func(string)) {
	if targetTag == "" {
		statusFn("Non renseigne")
		return
	}
	time.Sleep(3 * time.Second)
	found := false

	for _, g := range s.State.Guilds {
		for _, m := range g.Members {
			tag := m.User.Username + "#" + m.User.Discriminator
			if m.User.Discriminator == "0" {
				tag = m.User.Username
			}
			if tag == targetTag || m.User.Username == targetTag {
				found = true
				break
			}
		}
	}

	if found {
		statusFn("Utilisateur detecte !")
		logFn("Utilisateur cible (" + targetTag + ") trouve sur les serveurs.")
	} else {
		statusFn("Introuvable")
		logFn("Utilisateur cible non trouve, le bot continue de tourner.")
	}
}

func stopBot(logFn func(string), statusFn func(bool)) {
	safeRun(func() {
		state.Mutex.Lock()
		if !state.Running || state.Session == nil {
			state.Mutex.Unlock()
			logFn("Le bot est deja arrete.")
			return
		}
		err := state.Session.Close()
		if err != nil {
			logFn("Erreur lors de l'arret du bot.")
		}
		state.Session = nil
		state.Running = false
		state.Mutex.Unlock()
		logFn("Le bot a ete arrete manuellement.")
		statusFn(false)
	})
}

func setPresence(s *discordgo.Session, logFn func(string)) {
	t := discordgo.ActivityTypeGame
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
		logFn("Erreur lors de la mise a jour du statut du bot.")
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
		logFn("Message de " + m.Author.Username + " : " + m.Content)
	}

	// Correction de la variable de syntaxe ici : aSupprimer
	if cfg.Features.SimpleModeration {
		insultes := []string{"merde", "con", "salaud", "connard"}
		contenu := strings.ToLower(m.Content)
		aSupprimer := false
		for _, mot := range insultes {
			if strings.Contains(contenu, mot) {
				aSupprimer = true
				break
			}
		}
		if aSupprimer {
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			logFn("Moderation : Message supprime de " + m.Author.Username)
			s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+", merci de rester poli.")
			return
		}
	}

	if cfg.Features.AutoReaction {
		if strings.ToLower(m.Content) == "salut" || strings.ToLower(m.Content) == "bonjour" {
			s.ChannelMessageSend(m.ChannelID, "Bonjour "+m.Author.Mention()+" ! Heureux de te voir ici.")
		}
	}

	if cfg.Features.PrefixCommands && strings.HasPrefix(m.Content, cfg.Prefix) {
		content := strings.TrimPrefix(m.Content, cfg.Prefix)
		parts := strings.Fields(content)
		if len(parts) == 0 {
			return
		}
		cmd := strings.ToLower(parts[0])

		if cfg.Features.Ping && cmd == "ping" {
			_, _ = s.ChannelMessageSendReply(m.ChannelID, "Pong !", m.Reference())
		}
		if cfg.Features.Help && cmd == "help" {
			_, _ = s.ChannelMessageSendReply(m.ChannelID, "Voici les commandes actives :\n"+listCommands(), m.Reference())
		}
		
		for _, c := range cfg.Commands {
			if m.Content == c.Trigger || cmd == strings.ToLower(strings.TrimPrefix(c.Trigger, cfg.Prefix)) {
				_, _ = s.ChannelMessageSendReply(m.ChannelID, c.Reply, m.Reference())
			}
		}
	}
}

func handleNewMember(s *discordgo.Session, h *discordgo.GuildMemberAdd, logFn func(string)) {
	if cfg.Features.Welcome {
		guild, err := s.State.Guild(h.GuildID)
		if err == nil {
			txtChannelID := guild.SystemChannelID
			if txtChannelID == "" {
				channels, _ := s.GuildChannels(h.GuildID)
				for _, ch := range channels {
					if ch.Type == discordgo.ChannelTypeGuildText {
						txtChannelID = ch.ID
						break
					}
				}
			}
			if txtChannelID != "" {
				s.ChannelMessageSend(txtChannelID, "Bienvenue sur le serveur " + h.User.Mention() + " !")
				logFn("Bienvenue envoyee pour " + h.User.Username)
			}
		}
	}

	if cfg.Features.DmWelcome {
		channel, err := s.UserChannelCreate(h.User.ID)
		if err == nil {
			s.ChannelMessageSend(channel.ID, "Bienvenue ! Merci de rejoindre notre serveur public.")
			logFn("DM de bienvenue envoye a " + h.User.Username)
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
	for _, c := range cfg.Commands {
		list = append(list, c.Trigger)
	}
	return strings.Join(list, "\n")
}

func checkToken(token string) (bool, string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return false, ""
	}
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return false, ""
	}
	u, err := s.User("@me")
	if err == nil && u != nil {
		return true, u.Username
	}
	return false, ""
}

func main() {
	log.SetOutput(os.Stdout)

	a := app.NewWithID("bot.discord.creator")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Createur de Bot Discord reel")
	w.Resize(fyne.NewSize(1200, 800))

	logBuffer := ""
	logMutex := sync.Mutex{}

	logsWidget := widget.NewMultiLineEntry()
	logsWidget.SetPlaceHolder("Les evenements reels de ton bot apparaitront ici lors des actions des utilisateurs...")
	logsWidget.Disable()

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
		logsWidget.SetText(logBuffer)
		logsWidget.CursorRow = len(strings.Split(logBuffer, "\n")) - 1
	}

	// Utilisation d'un seul widget de statut avec émojis pour éviter les bugs Fyne
	statusLabel := widget.NewLabel("Statut : Hors ligne")

	setOnline := func(on bool) {
		if on {
			statusLabel.SetText("Statut : En ligne")
		} else {
			statusLabel.SetText("Statut : Hors ligne")
		}
		statusLabel.Refresh()
	}

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("Colle ton token Discord valide ici")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Nom du bot charge automatiquement")

	prefixEntry := widget.NewEntry()
	prefixEntry.SetText("!")

	targetUserEntry := widget.NewEntry()
	targetUserEntry.SetPlaceHolder("Ex: MonPseudo")
	
	targetStatusLabel := widget.NewLabel("En attente du lancement")

	rssEntry := widget.NewEntry()
	rssEntry.SetPlaceHolder("Flux RSS optionnel")

	statusTextEntry := widget.NewEntry()
	statusTextEntry.SetText("Actif et connecte !")

	statusTypeSelect := widget.NewSelect([]string{"PLAYING", "WATCHING", "LISTENING", "COMPETING"}, func(string) {})
	statusTypeSelect.SetSelected("PLAYING")

	tokenStatusLabel := widget.NewLabel("Aucun token verifie")
	checkTokenBtn := widget.NewButton("1. Verifier le token via requete REST", func() {
		t := strings.TrimSpace(tokenEntry.Text)
		if t == "" {
			tokenStatusLabel.SetText("Token vide")
			return
		}
		tokenStatusLabel.SetText("Verification reelle sur l'API Discord...")
		go func() {
			ok, botName := checkToken(t)
			if ok {
				tokenStatusLabel.SetText("Token valide ! Pret a l'action.")
				if botName != "" {
					nameEntry.SetText(botName)
				}
			} else {
				tokenStatusLabel.SetText("Token invalide ou rejete par Discord.")
			}
		}()
	})

	featureCheck := func(label string, get func() bool, set func(bool)) *widget.Check {
		c := widget.NewCheck(label, set)
		c.SetChecked(get())
		return c
	}

	featuresBtn := widget.NewButtonWithIcon("Configurer les fonctionnalites actives", theme.SettingsIcon(), func() {
		contentBox := container.NewVBox(
			widget.NewLabelWithStyle("Options generales", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			featureCheck("Repondre aux commandes avec prefixe (ex: !aide)", func() bool { return cfg.Features.PrefixCommands }, func(b bool) { cfg.Features.PrefixCommands = b }),
			featureCheck("Activer les logs dans la console", func() bool { return cfg.Features.Logs }, func(b bool) { cfg.Features.Logs = b }),
			featureCheck("Activer l'Auto-Reaction (Repondre a bonjour)", func() bool { return cfg.Features.AutoReaction }, func(b bool) { cfg.Features.AutoReaction = b }),
			featureCheck("Moderation Simple (Supprime les gros mots automatiquement)", func() bool { return cfg.Features.SimpleModeration }, func(b bool) { cfg.Features.SimpleModeration = b }),
			featureCheck("Envoyer un message de bienvenue sur le serveur", func() bool { return cfg.Features.Welcome }, func(b bool) { cfg.Features.Welcome = b }),
			featureCheck("Envoyer un DM prive de bienvenue aux nouveaux", func() bool { return cfg.Features.DmWelcome }, func(b bool) { cfg.Features.DmWelcome = b }),
			featureCheck("Commande !ping active", func() bool { return cfg.Features.Ping }, func(b bool) { cfg.Features.Ping = b }),
			featureCheck("Commande !help active", func() bool { return cfg.Features.Help }, func(b bool) { cfg.Features.Help = b }),
		)

		scrollContent := container.NewVScroll(contentBox)
		scrollContent.SetMinSize(fyne.NewSize(500, 400))

		dialog.ShowCustom("Gerer les Fonctionnalites", "Sauvegarder et Fermer", scrollContent, w)
	})

	cfg.Features.Ping = true
	cfg.Features.Help = true
	cfg.Features.Logs = true
	cfg.Features.PrefixCommands = true
	cfg.Features.SimpleModeration = true
	cfg.Features.AutoReaction = true

	cmdNameEntry := widget.NewEntry()
	cmdNameEntry.SetPlaceHolder("Ex: Dire Bonjour")
	cmdTriggerEntry := widget.NewEntry()
	cmdTriggerEntry.SetPlaceHolder("Ex: !bonjour")
	cmdReplyEntry := widget.NewMultiLineEntry()
	cmdReplyEntry.SetPlaceHolder("La reponse automatique")

	commandsList := widget.NewMultiLineEntry()
	commandsList.Disable()
	commandsList.SetPlaceHolder("Aucune commande personnalisee.")

	refreshCommandsView := func() {
		if len(cfg.Commands) == 0 {
			commandsList.SetText("Aucune commande personnalisee.")
			return
		}
		var lines []string
		for i, c := range cfg.Commands {
			lines = append(lines, fmt.Sprintf("%d. %s : Taper '%s' -> '%s'", i+1, c.Name, c.Trigger, c.Reply))
		}
		commandsList.SetText(strings.Join(lines, "\n"))
	}

	addCmdBtn := widget.NewButtonWithIcon("Ajouter la commande", theme.DocumentCreateIcon(), func() {
		name := strings.TrimSpace(cmdNameEntry.Text)
		trig := strings.TrimSpace(cmdTriggerEntry.Text)
		rep := strings.TrimSpace(cmdReplyEntry.Text)
		if name == "" || trig == "" || rep == "" {
			return
		}
		cfg.Commands = append(cfg.Commands, CustomCommand{Name: name, Trigger: trig, Reply: rep})
		cmdNameEntry.SetText("")
		cmdTriggerEntry.SetText("")
		cmdReplyEntry.SetText("")
		refreshCommandsView()
	})

	startBtn := widget.NewButtonWithIcon("DEMARRER LE BOT", theme.MediaPlayIcon(), func() {
		cfg.Token = tokenEntry.Text
		cfg.Name = nameEntry.Text
		cfg.TargetUserTag = targetUserEntry.Text
		cfg.RSSFeedURL = rssEntry.Text
		if strings.TrimSpace(prefixEntry.Text) == "" {
			cfg.Prefix = "!"
		} else {
			cfg.Prefix = prefixEntry.Text
		}
		cfg.StatusText = statusTextEntry.Text
		cfg.StatusType = statusTypeSelect.Selected

		go startBot(appendLog, setOnline, func(status string) {
			targetStatusLabel.SetText(status)
		})
	})
	startBtn.Importance = widget.HighImportance

	stopBtn := widget.NewButtonWithIcon("ARRETER", theme.MediaStopIcon(), func() {
		go stopBot(appendLog, setOnline)
	})
	stopBtn.Importance = widget.DangerImportance

	topBar := container.NewHBox(
		statusLabel,
		layout.NewSpacer(),
		widget.NewLabel("Gestionnaire Reel de Bot Discord"),
	)

	leftColBox := container.NewVBox(
		widget.NewLabelWithStyle("1. Connexion au Bot", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tokenEntry,
		container.NewHBox(checkTokenBtn, tokenStatusLabel),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("2. Parametres de ciblage", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Nom du bot"), nameEntry),
			container.NewVBox(widget.NewLabel("Prefixe general"), prefixEntry),
		),
		
		widget.NewLabel("Cibler un utilisateur par son pseudo"),
		container.NewGridWithColumns(2, targetUserEntry, targetStatusLabel),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("3. Presence", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Texte d'activite"), statusTextEntry),
			container.NewVBox(widget.NewLabel("Type"), statusTypeSelect),
		),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("4. Gestion du moteur", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewPadded(featuresBtn),
	)
	leftCol := container.NewVScroll(leftColBox)

	rightColBox := container.NewVBox(
		widget.NewLabelWithStyle("Creer une commande reelle", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Nom repere"), cmdNameEntry),
			container.NewVBox(widget.NewLabel("Texte declencheur"), cmdTriggerEntry),
		),
		widget.NewLabel("Reponse du bot"),
		cmdReplyEntry,
		container.NewPadded(addCmdBtn),
		commandsList,
		widget.NewSeparator(),
		
		widget.NewLabelWithStyle("Console d'evenements en direct", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		logsWidget,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, stopBtn, startBtn),
	)
	rightCol := container.NewVScroll(rightColBox)

	mainSplit := container.NewHSplit(leftCol, rightCol)
	mainSplit.Offset = 0.5

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
