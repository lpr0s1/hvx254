@echo off
setlocal

set DOWNLOADS=%USERPROFILE%\Downloads
set BOTDIR=%DOWNLOADS%\bot_discord_gui

echo Creation du dossier...
mkdir "%BOTDIR%" 2>nul
cd /d "%BOTDIR%"

echo Installation de Chocolatey...
powershell -NoProfile -ExecutionPolicy Bypass -Command "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"

if exist "%ProgramData%\chocolatey\bin\refreshenv.cmd" (
  call "%ProgramData%\chocolatey\bin\refreshenv.cmd"
)

echo Installation de Node.js...
choco install nodejs -y

if exist "%ProgramData%\chocolatey\bin\refreshenv.cmd" (
  call "%ProgramData%\chocolatey\bin\refreshenv.cmd"
)

echo Verification de node et npm...
node -v
npm -v

echo Initialisation du projet...
npm init -y

echo Installation des dependances...
npm install electron@28.2.0 discord.js

echo Creation de main.js...

> main.js echo const { app, BrowserWindow, ipcMain } = require("electron")
>> main.js echo const path = require("path")
>> main.js echo const { Client, GatewayIntentBits, ActivityType } = require("discord.js")
>> main.js echo let win = null
>> main.js echo let client = null
>> main.js echo let botStarted = false
>> main.js echo let config = {
>> main.js echo   token: "",
>> main.js echo   name: "",
>> main.js echo   prefix: "!",
>> main.js echo   statusText: "En ligne",
>> main.js echo   statusType: "PLAYING",
>> main.js echo   description: "",
>> main.js echo   features: {
>> main.js echo     ping: true,
>> main.js echo     help: true,
>> main.js echo     news: false,
>> main.js echo     faq: false,
>> main.js echo     bienvenue: false,
>> main.js echo     auRevoir: false,
>> main.js echo     autoReponse: false,
>> main.js echo     blagues: false,
>> main.js echo     citations: false,
>> main.js echo     moderationSimple: false,
>> main.js echo     antiSpam: false,
>> main.js echo     logs: true,
>> main.js echo     compteurMessages: false,
>> main.js echo     compteurMembres: false,
>> main.js echo     reactionAuto: false,
>> main.js echo     commandesSlash: false,
>> main.js echo     commandesPrefix: true,
>> main.js echo     messagePriveBienvenue: false,
>> main.js echo     messagePriveInfo: false,
>> main.js echo     rappel: false,
>> main.js echo     sondages: false,
>> main.js echo     rolesAuto: false,
>> main.js echo     rolesReaction: false,
>> main.js echo     musiquePlaceholder: false,
>> main.js echo     infoServeur: false,
>> main.js echo     infoUtilisateur: false,
>> main.js echo     tempsEnLigne: false,
>> main.js echo     statsJour: false,
>> main.js echo     statsSemaine: false,
>> main.js echo     statsMois: false,
>> main.js echo     messageAnnonce: false
>> main.js echo   },
>> main.js echo   customCommands: [],
>> main.js echo   logs: []
>> main.js echo }
>> main.js echo function createWindow() {
>> main.js echo   win = new BrowserWindow({
>> main.js echo     width: 1100,
>> main.js echo     height: 700,
>> main.js echo     resizable: true,
>> main.js echo     webPreferences: {
>> main.js echo       preload: path.join(__dirname, "preload.js")
>> main.js echo     },
>> main.js echo     backgroundColor: "#111827",
>> main.js echo     title: "Createur de bot Discord"
>> main.js echo   })
>> main.js echo   win.loadFile("index.html")
>> main.js echo }
>> main.js echo app.whenReady().then(() => {
>> main.js echo   createWindow()
>> main.js echo   app.on("activate", () => {
>> main.js echo     if (BrowserWindow.getAllWindows().length === 0) createWindow()
>> main.js echo   })
>> main.js echo })
>> main.js echo app.on("window-all-closed", () => {
>> main.js echo   if (process.platform !== "darwin") app.quit()
>> main.js echo })
>> main.js echo function sendLog(text) {
>> main.js echo   config.logs.push(text)
>> main.js echo   if (config.logs.length > 300) config.logs.shift()
>> main.js echo   if (win) win.webContents.send("log", text)
>> main.js echo }
>> main.js echo function setStatus() {
>> main.js echo   if (!client || !client.user) return
>> main.js echo   let type = ActivityType.Playing
>> main.js echo   if (config.statusType === "WATCHING") type = ActivityType.Watching
>> main.js echo   else if (config.statusType === "LISTENING") type = ActivityType.Listening
>> main.js echo   else if (config.statusType === "COMPETING") type = ActivityType.Competing
>> main.js echo   client.user.setPresence({
>> main.js echo     activities: [{ name: config.statusText, type }],
>> main.js echo     status: "online"
>> main.js echo   })
>> main.js echo }
>> main.js echo function handleMessage(message) {
>> main.js echo   if (message.author.bot) return
>> main.js echo   if (config.features.logs) sendLog("Message de " + message.author.tag + " : " + message.content)
>> main.js echo   if (config.features.moderationSimple) {
>> main.js echo     const badWords = ["insulte1", "insulte2"]
>> main.js echo     if (badWords.some(w => message.content.toLowerCase().includes(w))) {
>> main.js echo       message.delete().catch(() => {})
>> main.js echo       message.channel.send("Ce message a ete supprime.").catch(() => {})
>> main.js echo       return
>> main.js echo     }
>> main.js echo   }
>> main.js echo   if (config.features.reactionAuto) {
>> main.js echo     if (message.content.toLowerCase().includes("bonjour")) {
>> main.js echo       message.react("o").catch(() => {})
>> main.js echo     }
>> main.js echo   }
>> main.js echo   if (config.features.autoReponse) {
>> main.js echo     if (message.content.toLowerCase().includes("comment ca va")) {
>> main.js echo       message.reply("Je vais bien, merci.").catch(() => {})
>> main.js echo     }
>> main.js echo   }
>> main.js echo   if (config.features.blagues && message.content.toLowerCase().includes("blague")) {
>> main.js echo     message.reply("Pourquoi le bot traverse la route ? Pour aller sur un autre serveur.").catch(() => {})
>> main.js echo   }
>> main.js echo   if (config.features.citations && message.content.toLowerCase().includes("citation")) {
>> main.js echo     message.reply("La patience est la cle de beaucoup de choses.").catch(() => {})
>> main.js echo   }
>> main.js echo   if (config.features.commandesPrefix && message.content.startsWith(config.prefix)) {
>> main.js echo     const args = message.content.slice(config.prefix.length).trim().split(/\s+/)
>> main.js echo     const cmd = args.shift().toLowerCase()
>> main.js echo     if (config.features.ping && cmd === "ping") {
>> main.js echo       message.reply("Pong !").catch(() => {})
>> main.js echo     }
>> main.js echo     if (config.features.help && cmd === "help") {
>> main.js echo       message.reply("Commandes: " + listBasicCommands()).catch(() => {})
>> main.js echo     }
>> main.js echo     if (config.features.news && cmd === "news") {
>> main.js echo       message.reply("Pas de news pour le moment.").catch(() => {})
>> main.js echo     }
>> main.js echo     if (config.features.faq && cmd === "faq") {
>> main.js echo       message.reply("Pas de FAQ pour le moment.").catch(() => {})
>> main.js echo     }
>> main.js echo     if (config.features.infoServeur && cmd === "server") {
>> main.js echo       message.reply("Nom du serveur: " + message.guild.name).catch(() => {})
>> main.js echo     }
>> main.js echo     if (config.features.infoUtilisateur && cmd === "me") {
>> main.js echo       message.reply("Tu es " + message.author.tag).catch(() => {})
>> main.js echo     }
>> main.js echo     config.customCommands.forEach(c => {
>> main.js echo       if (message.content === c.trigger) {
>> main.js echo         message.reply(c.reply).catch(() => {})
>> main.js echo       }
>> main.js echo     })
>> main.js echo   }
>> main.js echo }
>> main.js echo function listBasicCommands() {
>> main.js echo   const list = []
>> main.js echo   if (config.features.ping) list.push(config.prefix + "ping")
>> main.js echo   if (config.features.help) list.push(config.prefix + "help")
>> main.js echo   if (config.features.news) list.push(config.prefix + "news")
>> main.js echo   if (config.features.faq) list.push(config.prefix + "faq")
>> main.js echo   if (config.features.infoServeur) list.push(config.prefix + "server")
>> main.js echo   if (config.features.infoUtilisateur) list.push(config.prefix + "me")
>> main.js echo   config.customCommands.forEach(c => list.push(c.trigger))
>> main.js echo   return list.join(", ")
>> main.js echo }
>> main.js echo function handleMemberJoin(member) {
>> main.js echo   if (config.features.bienvenue) {
>> main.js echo     const channel = member.guild.systemChannel || member.guild.channels.cache.find(c => c.isTextBased())
>> main.js echo     if (channel) channel.send("Bienvenue " + member.user.username + " !").catch(() => {})
>> main.js echo   }
>> main.js echo   if (config.features.messagePriveBienvenue) {
>> main.js echo     member.send("Bienvenue sur le serveur.").catch(() => {})
>> main.js echo   }
>> main.js echo }
>> main.js echo function handleMemberLeave(member) {
>> main.js echo   if (config.features.auRevoir) {
>> main.js echo     const channel = member.guild.systemChannel || member.guild.channels.cache.find(c => c.isTextBased())
>> main.js echo     if (channel) channel.send(member.user.username + " a quitte le serveur.").catch(() => {})
>> main.js echo   }
>> main.js echo }
>> main.js echo ipcMain.handle("check-token", async (event, token) => {
>> main.js echo   const testClient = new Client({ intents: [GatewayIntentBits.Guilds] })
>> main.js echo   try {
>> main.js echo     await testClient.login(token)
>> main.js echo     await testClient.destroy()
>> main.js echo     return { ok: true }
>> main.js echo   } catch (e) {
>> main.js echo     return { ok: false }
>> main.js echo   }
>> main.js echo })
>> main.js echo ipcMain.on("update-config", (event, data) => {
>> main.js echo   config = data
>> main.js echo })
>> main.js echo ipcMain.handle("start-bot", async () => {
>> main.js echo   if (botStarted) return { started: true }
>> main.js echo   if (!config.token) return { started: false, error: "Pas de token" }
>> main.js echo   client = new Client({
>> main.js echo     intents: [
>> main.js echo       GatewayIntentBits.Guilds,
>> main.js echo       GatewayIntentBits.GuildMessages,
>> main.js echo       GatewayIntentBits.MessageContent,
>> main.js echo       GatewayIntentBits.GuildMembers
>> main.js echo     ]
>> main.js echo   })
>> main.js echo   client.on("ready", () => {
>> main.js echo     botStarted = true
>> main.js echo     sendLog("Bot connecte en tant que " + client.user.tag)
>> main.js echo     setStatus()
>> main.js echo     if (win) win.webContents.send("bot-status", { online: true })
>> main.js echo   })
>> main.js echo   client.on("messageCreate", message => {
>> main.js echo     handleMessage(message)
>> main.js echo   })
>> main.js echo   client.on("guildMemberAdd", member => {
>> main.js echo     handleMemberJoin(member)
>> main.js echo   })
>> main.js echo   client.on("guildMemberRemove", member => {
>> main.js echo     handleMemberLeave(member)
>> main.js echo   })
>> main.js echo   try {
>> main.js echo     await client.login(config.token)
>> main.js echo     return { started: true }
>> main.js echo   } catch (e) {
>> main.js echo     sendLog("Erreur de demarrage du bot")
>> main.js echo     return { started: false, error: "Erreur de demarrage" }
>> main.js echo   }
>> main.js echo })
>> main.js echo ipcMain.handle("stop-bot", async () => {
>> main.js echo   if (!botStarted || !client) return { stopped: true }
>> main.js echo   await client.destroy()
>> main.js echo   botStarted = false
>> main.js echo   sendLog("Bot arrete")
>> main.js echo   if (win) win.webContents.send("bot-status", { online: false })
>> main.js echo   return { stopped: true }
>> main.js echo })
>> main.js echo ipcMain.handle("get-logs", () => {
>> main.js echo   return config.logs
>> main.js echo })

echo Creation de preload.js...

> preload.js echo const { contextBridge, ipcRenderer } = require("electron")
>> preload.js echo contextBridge.exposeInMainWorld("botApi", {
>> preload.js echo   checkToken: (token) => ipcRenderer.invoke("check-token", token),
>> preload.js echo   updateConfig: (config) => ipcRenderer.send("update-config", config),
>> preload.js echo   startBot: () => ipcRenderer.invoke("start-bot"),
>> preload.js echo   stopBot: () => ipcRenderer.invoke("stop-bot"),
>> preload.js echo   getLogs: () => ipcRenderer.invoke("get-logs"),
>> preload.js echo   onLog: (cb) => ipcRenderer.on("log", (e, msg) => cb(msg)),
>> preload.js echo   onStatus: (cb) => ipcRenderer.on("bot-status", (e, s) => cb(s))
>> preload.js echo })

echo Mise a jour de package.json...

> package.json.tmp echo {
for /f "usebackq delims=" %%A in (`type package.json`) do >> package.json.tmp echo %%A
>> package.json.tmp echo , "main": "main.js", "scripts": { "start": "electron ." }
>> package.json.tmp echo }
move /y package.json.tmp package.json >nul

echo Creation de index.html...

> index.html echo ^<!DOCTYPE html^>
>> index.html echo ^<html lang="fr"^>
>> index.html echo ^<head^>
>> index.html echo ^<meta charset="UTF-8"^>
>> index.html echo ^<title^>Createur de bot Discord^</title^>
>> index.html echo ^<style^>
>> index.html echo body { margin: 0; font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #020617; color: #e5e7eb; }
>> index.html echo .app { display: flex; flex-direction: column; height: 100vh; }
>> index.html echo .header { padding: 12px 20px; background: #020617; border-bottom: 1px solid #1f2937; display: flex; align-items: center; justify-content: space-between; }
>> index.html echo .title { font-size: 16px; font-weight: 600; color: #e5e7eb; }
>> index.html echo .status-pill { padding: 4px 10px; border-radius: 999px; font-size: 12px; display: inline-flex; align-items: center; gap: 6px; }
>> index.html echo .status-on { background: #16a34a33; color: #4ade80; }
>> index.html echo .status-off { background: #b91c1c33; color: #fca5a5; }
>> index.html echo .status-dot { width: 8px; height: 8px; border-radius: 999px; }
>> index.html echo .status-dot-on { background: #22c55e; }
>> index.html echo .status-dot-off { background: #ef4444; }
>> index.html echo .content { display: grid; grid-template-columns: 1.3fr 1fr; gap: 12px; padding: 12px 16px 16px; height: calc(100vh - 48px); box-sizing: border-box; }
>> index.html echo .card { background: #020617; border-radius: 10px; border: 1px solid #1f2937; padding: 12px 14px; box-sizing: border-box; }
>> index.html echo .card-title { font-size: 13px; font-weight: 600; margin-bottom: 8px; color: #e5e7eb; }
>> index.html echo .field { margin-bottom: 8px; }
>> index.html echo .label { font-size: 12px; color: #9ca3af; margin-bottom: 3px; }
>> index.html echo .input, .select, .textarea { width: 100%; background: #020617; border-radius: 6px; border: 1px solid #1f2937; color: #e5e7eb; font-size: 12px; padding: 7px 8px; box-sizing: border-box; outline: none; }
>> index.html echo .input:focus, .select:focus, .textarea:focus { border-color: #6366f1; }
>> index.html echo .textarea { resize: vertical; min-height: 60px; max-height: 140px; }
>> index.html echo .row { display: flex; gap: 8px; }
>> index.html echo .row > .field { flex: 1; }
>> index.html echo .btn { border: none; border-radius: 6px; padding: 7px 10px; font-size: 12px; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; gap: 6px; }
>> index.html echo .btn-main { background: #4f46e5; color: white; }
>> index.html echo .btn-main:hover { background: #6366f1; }
>> index.html echo .btn-ghost { background: #020617; color: #e5e7eb; border: 1px solid #1f2937; }
>> index.html echo .btn-ghost:hover { border-color: #4b5563; }
>> index.html echo .btn-danger { background: #b91c1c; color: white; }
>> index.html echo .btn-danger:hover { background: #dc2626; }
>> index.html echo .btn-small { padding: 5px 8px; font-size: 11px; }
>> index.html echo .token-row { display: flex; gap: 8px; }
>> index.html echo .token-row .input { flex: 1; }
>> index.html echo .pill { padding: 4px 8px; border-radius: 999px; font-size: 11px; border: 1px solid #1f2937; color: #9ca3af; display: inline-flex; align-items: center; gap: 6px; }
>> index.html echo .pill-dot { width: 6px; height: 6px; border-radius: 999px; background: #4b5563; }
>> index.html echo .pill-ok { color: #4ade80; border-color: #16a34a66; }
>> index.html echo .pill-ok .pill-dot { background: #22c55e; }
>> index.html echo .pill-bad { color: #fca5a5; border-color: #b91c1c66; }
>> index.html echo .pill-bad .pill-dot { background: #ef4444; }
>> index.html echo .features-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 4px 10px; max-height: 210px; overflow-y: auto; padding-right: 4px; }
>> index.html echo .feature-item { display: flex; align-items: center; gap: 6px; font-size: 12px; color: #d1d5db; }
>> index.html echo .feature-item input { width: 14px; height: 14px; }
>> index.html echo .logs-box { background: #020617; border-radius: 8px; border: 1px solid #1f2937; padding: 8px; font-size: 11px; color: #9ca3af; height: 180px; overflow-y: auto; white-space: pre-wrap; }
>> index.html echo .commands-list { max-height: 120px; overflow-y: auto; border-radius: 6px; border: 1px solid #1f2937; padding: 6px 8px; font-size: 11px; color: #9ca3af; }
>> index.html echo .command-item { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
>> index.html echo .command-name { color: #e5e7eb; }
>> index.html echo .command-trigger { color: #9ca3af; font-size: 10px; }
>> index.html echo .footer-buttons { display: flex; justify-content: flex-end; gap: 8px; margin-top: 8px; }
>> index.html echo .hint { font-size: 11px; color: #6b7280; margin-top: 4px; }
>> index.html echo .divider { height: 1px; background: #111827; margin: 6px 0; }
>> index.html echo ^</style^>
>> index.html echo ^</head^>
>> index.html echo ^<body^>
>> index.html echo ^<div class="app"^>
>> index.html echo ^<div class="header"^>
>> index.html echo ^<div class="title"^>Createur de bot Discord^</div^>
>> index.html echo ^<div id="botStatus" class="status-pill status-off"^>
>> index.html echo ^<div class="status-dot status-dot-off"^>^</div^>
>> index.html echo ^<span^>Hors ligne^</span^>
>> index.html echo ^</div^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="content"^>
>> index.html echo ^<div class="card"^>
>> index.html echo ^<div class="card-title"^>Connexion et informations^</div^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>As tu cree ton bot sur le site Discord Developer ?^</div^>
>> index.html echo ^<div class="row"^>
>> index.html echo ^<button id="openDev" class="btn btn-ghost btn-small"^>Ouvrir Discord Developer^</button^>
>> index.html echo ^<span class="hint"^>Va sur le site, cree une application, ajoute un bot, copie le token.^</span^>
>> index.html echo ^</div^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>Token du bot^</div^>
>> index.html echo ^<div class="token-row"^>
>> index.html echo ^<input id="tokenInput" class="input" type="password" placeholder="Colle ici le token du bot"^>
>> index.html echo ^<button id="checkTokenBtn" class="btn btn-main btn-small"^>Verifier^</button^>
>> index.html echo ^</div^>
>> index.html echo ^<div id="tokenStatus" class="hint"^>En attente de verification.^</div^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="divider"^>^</div^>
>> index.html echo ^<div class="row"^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>Nom du bot^</div^>
>> index.html echo ^<input id="nameInput" class="input" placeholder="Nom d affichage du bot"^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>Prefixe des commandes^</div^>
>> index.html echo ^<input id="prefixInput" class="input" value="!"^>
>> index.html echo ^</div^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="row"^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>Texte de statut^</div^>
>> index.html echo ^<input id="statusTextInput" class="input" value="En ligne"^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>Type de statut^</div^>
>> index.html echo ^<select id="statusTypeSelect" class="select"^>
>> index.html echo ^<option value="PLAYING"^>Joue a^</option^>
>> index.html echo ^<option value="WATCHING"^>Regarde^</option^>
>> index.html echo ^<option value="LISTENING"^>Ecoute^</option^>
>> index.html echo ^<option value="COMPETING"^>Participe a^</option^>
>> index.html echo ^</select^>
>> index.html echo ^</div^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>Description du bot^</div^>
>> index.html echo ^<textarea id="descriptionInput" class="textarea" placeholder="Decris a quoi sert ton bot"^>^</textarea^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="footer-buttons"^>
>> index.html echo ^<span id="saveInfoStatus" class="hint"^>Les infos seront prises en compte au demarrage du bot.^</span^>
>> index.html echo ^</div^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="card"^>
>> index.html echo ^<div class="card-title"^>Options du bot^</div^>
>> index.html echo ^<div class="features-grid"^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="ping" checked^>^<span^>Repondre a ping^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="help" checked^>^<span^>Commande d aide^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="news"^>^<span^>Commande news^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="faq"^>^<span^>Questions frequentes^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="bienvenue"^>^<span^>Message de bienvenue^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="auRevoir"^>^<span^>Message d au revoir^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="autoReponse"^>^<span^>Reponses automatiques simples^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="blagues"^>^<span^>Blagues^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="citations"^>^<span^>Citations^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="moderationSimple"^>^<span^>Moderation simple^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="antiSpam"^>^<span^>Anti spam basique^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="logs" checked^>^<span^>Logs du bot^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="compteurMessages"^>^<span^>Compteur de messages^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="compteurMembres"^>^<span^>Compteur de membres^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="reactionAuto"^>^<span^>Reactions automatiques^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="commandesSlash"^>^<span^>Commandes slash simples^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="commandesPrefix" checked^>^<span^>Commandes avec prefixe^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="messagePriveBienvenue"^>^<span^>Message prive de bienvenue^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="messagePriveInfo"^>^<span^>Message prive d info^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="rappel"^>^<span^>Rappels simples^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="sondages"^>^<span^>Sondages^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="rolesAuto"^>^<span^>Roles automatiques^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="rolesReaction"^>^<span^>Roles par reaction^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="musiquePlaceholder"^>^<span^>Musique simple^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="infoServeur"^>^<span^>Infos serveur^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="infoUtilisateur"^>^<span^>Infos utilisateur^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="tempsEnLigne"^>^<span^>Temps en ligne du bot^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="statsJour"^>^<span^>Stats du jour^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="statsSemaine"^>^<span^>Stats de la semaine^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="statsMois"^>^<span^>Stats du mois^</span^>^</label^>
>> index.html echo ^<label class="feature-item"^>^<input type="checkbox" data-feature="messageAnnonce"^>^<span^>Messages d annonce^</span^>^</label^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="hint"^>Coche ce que ton bot doit faire.^</div^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="card"^>
>> index.html echo ^<div class="card-title"^>Commandes personnalisees^</div^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>Nom de la commande^</div^>
>> index.html echo ^<input id="cmdNameInput" class="input" placeholder="Ex: help, news, info"^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>Texte a taper pour la declencher^</div^>
>> index.html echo ^<input id="cmdTriggerInput" class="input" placeholder="Ex: /help ou !help"^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>Reponse du bot^</div^>
>> index.html echo ^<textarea id="cmdReplyInput" class="textarea" placeholder="Ce que le bot doit repondre"^>^</textarea^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="footer-buttons"^>
>> index.html echo ^<button id="addCmdBtn" class="btn btn-ghost btn-small"^>Ajouter la commande^</button^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>Commandes creees^</div^>
>> index.html echo ^<div id="commandsList" class="commands-list"^>Aucune commande pour le moment.^</div^>
>> index.html echo ^</div^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="card"^>
>> index.html echo ^<div class="card-title"^>Logs et controle^</div^>
>> index.html echo ^<div class="field"^>
>> index.html echo ^<div class="label"^>Logs du bot^</div^>
>> index.html echo ^<div id="logsBox" class="logs-box"^>En attente de logs...^</div^>
>> index.html echo ^</div^>
>> index.html echo ^<div class="footer-buttons"^>
>> index.html echo ^<button id="refreshLogsBtn" class="btn btn-ghost btn-small"^>Actualiser les logs^</button^>
>> index.html echo ^<button id="stopBotBtn" class="btn btn-danger btn-small"^>Arreter le bot^</button^>
>> index.html echo ^<button id="startBotBtn" class="btn btn-main btn-small"^>Demarrer le bot^</button^>
>> index.html echo ^</div^>
>> index.html echo ^<div id="botActionStatus" class="hint"^>Le bot est pret a etre demarre.^</div^>
>> index.html echo ^</div^>
>> index.html echo ^</div^>
>> index.html echo ^<script^>
>> index.html echo const tokenInput = document.getElementById("tokenInput")
>> index.html echo const checkTokenBtn = document.getElementById("checkTokenBtn")
>> index.html echo const tokenStatus = document.getElementById("tokenStatus")
>> index.html echo const nameInput = document.getElementById("nameInput")
>> index.html echo const prefixInput = document.getElementById("prefixInput")
>> index.html echo const statusTextInput = document.getElementById("statusTextInput")
>> index.html echo const statusTypeSelect = document.getElementById("statusTypeSelect")
>> index.html echo const descriptionInput = document.getElementById("descriptionInput")
>> index.html echo const featureInputs = document.querySelectorAll("[data-feature]")
>> index.html echo const cmdNameInput = document.getElementById("cmdNameInput")
>> index.html echo const cmdTriggerInput = document.getElementById("cmdTriggerInput")
>> index.html echo const cmdReplyInput = document.getElementById("cmdReplyInput")
>> index.html echo const addCmdBtn = document.getElementById("addCmdBtn")
>> index.html echo const commandsList = document.getElementById("commandsList")
>> index.html echo const logsBox = document.getElementById("logsBox")
>> index.html echo const refreshLogsBtn = document.getElementById("refreshLogsBtn")
>> index.html echo const startBotBtn = document.getElementById("startBotBtn")
>> index.html echo const stopBotBtn = document.getElementById("stopBotBtn")
>> index.html echo const botStatus = document.getElementById("botStatus")
>> index.html echo const openDev = document.getElementById("openDev")
>> index.html echo const botActionStatus = document.getElementById("botActionStatus")
>> index.html echo let config = {
>> index.html echo   token: "",
>> index.html echo   name: "",
>> index.html echo   prefix: "!",
>> index.html echo   statusText: "En ligne",
>> index.html echo   statusType: "PLAYING",
>> index.html echo   description: "",
>> index.html echo   features: {},
>> index.html echo   customCommands: []
>> index.html echo }
>> index.html echo function updateConfigFromUI() {
>> index.html echo   config.token = tokenInput.value.trim()
>> index.html echo   config.name = nameInput.value.trim()
>> index.html echo   config.prefix = prefixInput.value.trim() || "!"
>> index.html echo   config.statusText = statusTextInput.value.trim() || "En ligne"
>> index.html echo   config.statusType = statusTypeSelect.value
>> index.html echo   config.description = descriptionInput.value.trim()
>> index.html echo   featureInputs.forEach(i => {
>> index.html echo     config.features[i.dataset.feature] = i.checked
>> index.html echo   })
>> index.html echo   window.botApi.updateConfig(config)
>> index.html echo }
>> index.html echo function renderCommands() {
>> index.html echo   if (!config.customCommands.length) {
>> index.html echo     commandsList.textContent = "Aucune commande pour le moment."
>> index.html echo     return
>> index.html echo   }
>> index.html echo   commandsList.innerHTML = ""
>> index.html echo   config.customCommands.forEach((c, idx) => {
>> index.html echo     const div = document.createElement("div")
>> index.html echo     div.className = "command-item"
>> index.html echo     const left = document.createElement("div")
>> index.html echo     left.innerHTML = "^<div class='command-name'^>" + c.name + "^</div^>^<div class='command-trigger'^>" + c.trigger + "^</div^>"
>> index.html echo     const right = document.createElement("button")
>> index.html echo     right.className = "btn btn-ghost btn-small"
>> index.html echo     right.textContent = "Supprimer"
>> index.html echo     right.onclick = () => {
>> index.html echo       config.customCommands.splice(idx, 1)
>> index.html echo       updateConfigFromUI()
>> index.html echo       renderCommands()
>> index.html echo     }
>> index.html echo     div.appendChild(left)
>> index.html echo     div.appendChild(right)
>> index.html echo     commandsList.appendChild(div)
>> index.html echo   })
>> index.html echo }
>> index.html echo checkTokenBtn.onclick = async () => {
>> index.html echo   const token = tokenInput.value.trim()
>> index.html echo   if (!token) {
>> index.html echo     tokenStatus.textContent = "Colle d abord le token."
>> index.html echo     tokenStatus.className = "hint"
>> index.html echo     return
>> index.html echo   }
>> index.html echo   tokenStatus.textContent = "Verification en cours..."
>> index.html echo   tokenStatus.className = "hint"
>> index.html echo   const res = await window.botApi.checkToken(token)
>> index.html echo   if (res.ok) {
>> index.html echo     tokenStatus.innerHTML = "^<span class='pill pill-ok'^>^<span class='pill-dot'^>^</span^>Token valide^</span^>"
>> index.html echo   } else {
>> index.html echo     tokenStatus.innerHTML = "^<span class='pill pill-bad'^>^<span class='pill-dot'^>^</span^>Token invalide^</span^>"
>> index.html echo   }
>> index.html echo   updateConfigFromUI()
>> index.html echo }
>> index.html echo addCmdBtn.onclick = () => {
>> index.html echo   const name = cmdNameInput.value.trim()
>> index.html echo   const trigger = cmdTriggerInput.value.trim()
>> index.html echo   const reply = cmdReplyInput.value.trim()
>> index.html echo   if (!name || !trigger || !reply) return
>> index.html echo   config.customCommands.push({ name, trigger, reply })
>> index.html echo   cmdNameInput.value = ""
>> index.html echo   cmdTriggerInput.value = ""
>> index.html echo   cmdReplyInput.value = ""
>> index.html echo   updateConfigFromUI()
>> index.html echo   renderCommands()
>> index.html echo }
>> index.html echo featureInputs.forEach(i => {
>> index.html echo   i.onchange = () => {
>> index.html echo     updateConfigFromUI()
>> index.html echo   }
>> index.html echo })
>> index.html echo nameInput.oninput = updateConfigFromUI
>> index.html echo prefixInput.oninput = updateConfigFromUI
>> index.html echo statusTextInput.oninput = updateConfigFromUI
>> index.html echo statusTypeSelect.onchange = updateConfigFromUI
>> index.html echo descriptionInput.oninput = updateConfigFromUI
>> index.html echo startBotBtn.onclick = async () => {
>> index.html echo   updateConfigFromUI()
>> index.html echo   botActionStatus.textContent = "Demarrage du bot, attends un peu..."
>> index.html echo   const res = await window.botApi.startBot()
>> index.html echo   if (res.started) {
>> index.html echo     botActionStatus.textContent = "Bot demarre."
>> index.html echo   } else {
>> index.html echo     botActionStatus.textContent = res.error || "Impossible de demarrer le bot."
>> index.html echo   }
>> index.html echo }
>> index.html echo stopBotBtn.onclick = async () => {
>> index.html echo   botActionStatus.textContent = "Arret du bot..."
>> index.html echo   const res = await window.botApi.stopBot()
>> index.html echo   if (res.stopped) {
>> index.html echo     botActionStatus.textContent = "Bot arrete."
>> index.html echo   }
>> index.html echo }
>> index.html echo refreshLogsBtn.onclick = async () => {
>> index.html echo   const logs = await window.botApi.getLogs()
>> index.html echo   if (!logs.length) {
>> index.html echo     logsBox.textContent = "Aucun log pour le moment."
>> index.html echo     return
>> index.html echo   }
>> index.html echo   logsBox.textContent = logs.join("\n")
>> index.html echo   logsBox.scrollTop = logsBox.scrollHeight
>> index.html echo }
>> index.html echo window.botApi.onLog(msg => {
>> index.html echo   if (logsBox.textContent === "En attente de logs...") logsBox.textContent = ""
>> index.html echo   logsBox.textContent += (logsBox.textContent ? "\n" : "") + msg
>> index.html echo   logsBox.scrollTop = logsBox.scrollHeight
>> index.html echo })
>> index.html echo window.botApi.onStatus(s => {
>> index.html echo   if (s.online) {
>> index.html echo     botStatus.className = "status-pill status-on"
>> index.html echo     botStatus.innerHTML = "^<div class='status-dot status-dot-on'^>^</div^>^<span^>En ligne^</span^>"
>> index.html echo   } else {
>> index.html echo     botStatus.className = "status-pill status-off"
>> index.html echo     botStatus.innerHTML = "^<div class='status-dot status-dot-off'^>^</div^>^<span^>Hors ligne^</span^>"
>> index.html echo   }
>> index.html echo })
>> index.html echo openDev.onclick = () => {
>> index.html echo   window.open("https://discord.com/developers/applications", "_blank")
>> index.html echo }
>> index.html echo updateConfigFromUI()
>> index.html echo renderCommands()
>> index.html echo ^</script^>
>> index.html echo ^</body^>
>> index.html echo ^</html^>

echo Lancement de l interface...
npm run start

endlocal
pause
