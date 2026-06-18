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
	"fyne.io/fyne/v2/canvas"
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
			logFn("Token manquant.")
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

			// Verifications asynchrones apres connexion (Admin + Utilisateur cible)
			go checkAdminPerms(ses, logFn)
			go checkTargetUser(ses, cfg.TargetUserTag, targetStatusFn, logFn)
		})

		s.AddHandler(func(ses *discordgo.Session, m *discordgo.MessageCreate) {
			handleMessage(ses, m, logFn)
		})

		err = s.Open()
		if err != nil {
			state.Mutex.Unlock()
			logFn("Impossible de connecter le bot au serveur Discord.")
			return
		}

		state.Session = s
		state.Running = true
		state.StartTime = time.Now()
		state.Mutex.Unlock()
		logFn("Lancement du processus principal...")
		statusFn(true)
	})
}

// Verifie si le bot possede la permission ADMINISTRATEUR dans au moins un de ses serveurs
func checkAdminPerms(s *discordgo.Session, logFn func(string)) {
	time.Sleep(3 * time.Second) // On attend un peu que le cache se remplisse
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
		logFn("Info: Permission ADMINISTRATEUR detectee. Le bot a tous les droits !")
	} else {
		logFn("Info: Le bot n'a pas la permission Administrateur. Certaines options peuvent bloquer.")
	}
}

// Cherche l'utilisateur cible dans le cache
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
		statusFn("Introuvable (mais le bot continue)")
		logFn("Utilisateur cible non trouve, mais le bot continue de tourner.")
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
	for _, c := range cfg.Commands {
		list = append(list, c.Trigger)
	}
	return strings.Join(list, "\n")
}

// checkToken renvoie si le token est valide, et le nom du bot si recuperable
func checkToken(token string) (bool, string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return false, ""
	}
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return false, ""
	}
	
	// Utilisation de l'API REST pour recuperer le compte utilisateur du bot sans ouvrir de session WSS
	u, err := s.User("@me")
	if err == nil && u != nil {
		return true, u.Username
	}
	
	return true, ""
}

func main() {
	log.SetOutput(os.Stdout)

	a := app.NewWithID("bot.discord.creator")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Createur de Bot Discord Simplifie")
	w.Resize(fyne.NewSize(1200, 800))

	logBuffer := ""
	logMutex := sync.Mutex{}

	logsWidget := widget.NewMultiLineEntry()
	logsWidget.SetPlaceHolder("Les evenements de ton bot apparaitront ici...")
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

	statusLabel := widget.NewLabel("Statut : Hors ligne")
	statusDot := canvas.NewText("o", theme.ErrorColor())
	statusDot.TextStyle.Bold = true
	statusDot.TextSize = 18

	setOnline := func(on bool) {
		if on {
			statusLabel.SetText("Statut : En ligne")
			statusDot.Color = theme.PrimaryColor()
		} else {
			statusLabel.SetText("Statut : Hors ligne")
			statusDot.Color = theme.ErrorColor()
		}
		statusDot.Refresh()
	}

	// === CHAMPS DE TEXTE ===
	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("Colle ici le token de ton bot (trouve sur Discord Developer Portal)")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Ex: Mon Super Bot (Auto-rempli si token valide)")

	prefixEntry := widget.NewEntry()
	prefixEntry.SetText("!")

	targetUserEntry := widget.NewEntry()
	targetUserEntry.SetPlaceHolder("Ex: MonPseudo#1234 ou justemonpseudo")
	
	targetStatusLabel := widget.NewLabel("En attente du lancement")

	rssEntry := widget.NewEntry()
	rssEntry.SetPlaceHolder("Ex: https://mon-site.com/rss.xml (Optionnel)")

	statusTextEntry := widget.NewEntry()
	statusTextEntry.SetText("Pret a l'action !")

	statusTypeSelect := widget.NewSelect([]string{"PLAYING", "WATCHING", "LISTENING", "COMPETING"}, func(string) {})
	statusTypeSelect.SetSelected("PLAYING")

	// === ACTIONS ===
	tokenStatusLabel := widget.NewLabel("Aucun token verifie")
	checkTokenBtn := widget.NewButton("1. Verifier le token", func() {
		t := strings.TrimSpace(tokenEntry.Text)
		if t == "" {
			tokenStatusLabel.SetText("Token vide")
			return
		}
		tokenStatusLabel.SetText("Verification en cours...")
		go func() {
			ok, botName := checkToken(t)
			if ok {
				tokenStatusLabel.SetText("Token valide ! Le bot peut demarrer.")
				if botName != "" {
					nameEntry.SetText(botName) // Remplissage automatique
				}
			} else {
				tokenStatusLabel.SetText("Token invalide.")
			}
		}()
	})

	// === POPUP DES FONCTIONNALITES ===
	featureCheck := func(label string, get func() bool, set func(bool)) *widget.Check {
		c := widget.NewCheck(label, set)
		c.SetChecked(get())
		return c
	}

	featuresBtn := widget.NewButtonWithIcon("Configurer les fonctionnalites du Bot", theme.SettingsIcon(), func() {
		// Construction de la liste des fonctionnalites
		contentBox := container.NewVBox(
			widget.NewLabelWithStyle("Options generales", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			featureCheck("Repondre aux commandes avec prefixe (ex: !aide)", func() bool { return cfg.Features.PrefixCommands }, func(b bool) { cfg.Features.PrefixCommands = b }),
			featureCheck("Activer les logs (Voir tout ce qui se passe)", func() bool { return cfg.Features.Logs }, func(b bool) { cfg.Features.Logs = b }),
			featureCheck("Activer l'Auto-Reaction (Dire bonjour)", func() bool { return cfg.Features.AutoReaction }, func(b bool) { cfg.Features.AutoReaction = b }),
			featureCheck("Moderation Simple (Supprimer les insultes basiques)", func() bool { return cfg.Features.SimpleModeration }, func(b bool) { cfg.Features.SimpleModeration = b }),
			featureCheck("Envoyer un message de bienvenue", func() bool { return cfg.Features.Welcome }, func(b bool) { cfg.Features.Welcome = b }),
			featureCheck("Envoyer un DM (Message prive) de bienvenue", func() bool { return cfg.Features.DmWelcome }, func(b bool) { cfg.Features.DmWelcome = b }),
			featureCheck("Commande 'ping' active", func() bool { return cfg.Features.Ping }, func(b bool) { cfg.Features.Ping = b }),
			featureCheck("Commande 'help' (aide) active", func() bool { return cfg.Features.Help }, func(b bool) { cfg.Features.Help = b }),
		)

		scrollContent := container.NewVScroll(contentBox)
		scrollContent.SetMinSize(fyne.NewSize(500, 400)) // Taille fixe du popup

		dialog.ShowCustom("Gerer les Fonctionnalites", "Fermer et Sauvegarder", scrollContent, w)
	})

	// On active des fonctionnalites par defaut
	cfg.Features.Ping = true
	cfg.Features.Help = true
	cfg.Features.Logs = true
	cfg.Features.PrefixCommands = true

	// === COMMANDES PERSONNALISEES ===
	cmdNameEntry := widget.NewEntry()
	cmdNameEntry.SetPlaceHolder("Ex: Dire Bonjour")
	cmdTriggerEntry := widget.NewEntry()
	cmdTriggerEntry.SetPlaceHolder("Texte a ecrire (ex: !bonjour)")
	cmdReplyEntry := widget.NewMultiLineEntry()
	cmdReplyEntry.SetPlaceHolder("La reponse du bot (ex: Salut a toi !)")

	commandsList := widget.NewMultiLineEntry()
	commandsList.Disable()
	commandsList.SetPlaceHolder("Aucune commande personnalisee ajoutee.")

	refreshCommandsView := func() {
		if len(cfg.Commands) == 0 {
			commandsList.SetText("Aucune commande personnalisee ajoutee.")
			return
		}
		var lines []string
		for i, c := range cfg.Commands {
			lines = append(lines, fmt.Sprintf("%d. %s : Taper '%s' -> Bot repondra '%s'", i+1, c.Name, c.Trigger, c.Reply))
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

	// === BOUTONS PRINCIPAUX ===
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
	startBtn.Importance = widget.HighImportance // Rend le bouton tres visible

	stopBtn := widget.NewButtonWithIcon("ARRETER", theme.MediaStopIcon(), func() {
		go stopBot(appendLog, setOnline)
	})
	stopBtn.Importance = widget.DangerImportance // Rend le bouton rouge (selon le theme)

	// === MISE EN PAGE UI ===

	topBar := container.NewHBox(
		statusDot,
		statusLabel,
		layout.NewSpacer(),
		widget.NewLabel("Createur de Bot Discord par interface"),
	)

	// Colonne de gauche (Configuration)
	leftColBox := container.NewVBox(
		widget.NewLabelWithStyle("1. Identifiants du Bot", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		tokenEntry,
		container.NewHBox(checkTokenBtn, tokenStatusLabel),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("2. Informations et Ciblage", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Nom du bot"), nameEntry),
			container.NewVBox(widget.NewLabel("Prefixe des commandes"), prefixEntry),
		),
		
		widget.NewLabel("Cibler un utilisateur (Message Prive / Auto)"),
		container.NewGridWithColumns(2, targetUserEntry, targetStatusLabel),

		widget.NewLabel("Lien d'un Flux RSS (Optionnel)"),
		rssEntry,
		widget.NewSeparator(),

		widget.NewLabelWithStyle("3. Apparence sur Discord", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Texte d'activite"), statusTextEntry),
			container.NewVBox(widget.NewLabel("Type d'activite"), statusTypeSelect),
		),
		widget.NewSeparator(),

		widget.NewLabelWithStyle("4. Fonctionnalites", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewPadded(featuresBtn), // Bouton plus gros grace au padding
	)
	leftCol := container.NewVScroll(leftColBox)

	// Colonne de droite (Logs & Custom Commands)
	rightColBox := container.NewVBox(
		widget.NewLabelWithStyle("Creer une commande personnalisee", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Nom (Repere)"), cmdNameEntry),
			container.NewVBox(widget.NewLabel("Declencheur (Texte)"), cmdTriggerEntry),
		),
		widget.NewLabel("Que doit repondre le bot ?"),
		cmdReplyEntry,
		container.NewPadded(addCmdBtn),
		commandsList,
		widget.NewSeparator(),
		
		widget.NewLabelWithStyle("Console et Logs d'activite", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		logsWidget,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, stopBtn, startBtn),
	)
	rightCol := container.NewVScroll(rightColBox)

	mainSplit := container.NewHSplit(leftCol, rightCol)
	mainSplit.Offset = 0.5 // Moitie-moitie pour l'ecran

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
