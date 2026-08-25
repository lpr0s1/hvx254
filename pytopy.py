import socket
import threading
import random
import string
import os
import sys
import time
import signal
from art import *

def kjl(signum, frame):
    os.system('clear')
    os.system('cls')
    os.system('clear')
    sys.exit(0)
signal.signal(signal.SIGINT, kjl)
os.system("clear")
PYTOPY_VERSION = "1.0"
class User:
    def __init__(self, conn, addr, pseudo, uid):
        self.conn = conn
        self.addr = addr
        self.pseudo = pseudo
        self.uid = uid
class P2PNode:
    def __init__(self):
        self.host = "0.0.0.0"
        self.port = 9999
        self.users = {}
        self.lock = threading.Lock()
        self.running = True
        self.local_uid = None
        self.local_pseudo = None
        self.server_thread = None
        self.current_node_users = 0
    def generate_uid(self):
        return ''.join(random.choices(string.ascii_letters + string.digits, k=8))
    def choose_identity(self):
        tprint("pytopy")
        self.local_pseudo = input("Choisissez votre pseudo : ").strip()
        if not self.local_pseudo:
            self.local_pseudo = "Anonyme"
        self.local_uid = self.generate_uid()
        print(f"Votre ID est : {self.local_uid}")
        return self.local_pseudo, self.local_uid
    def start_server(self):
        try:
            self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            self.sock.bind((self.host, self.port))
            self.sock.listen(5)
            print(f"{self.local_pseudo} [{self.local_uid}] est maintenant le noeud.")
            print("En attente de connexions...")            
            while self.running:
                conn, addr = self.sock.accept()
                thread = threading.Thread(target=self.handle_client, args=(conn, addr), daemon=True)
                thread.start()
        except Exception as e:
            if self.running:
                print(f"Erreur serveur : {e}")
        finally:
            if not self.running:
                try:
                    self.sock.close()
                except:
                    pass
    def handle_client(self, conn, addr):
        try:
            data = conn.recv(1024).decode()
            if not data:
                return
            parts = data.strip().split(':')
            if len(parts) >= 3:
                pseudo = parts[0]
                uid = parts[1]
                msg_type = parts[2]
                if msg_type == "JOIN":
                    user = User(conn, addr, pseudo, uid)
                    with self.lock:
                        self.users[addr] = user
                        self.current_node_users += 1
                        print(f"-> [{pseudo}#{uid}] a rejoint le noeud.")
                        print(f"Il y a {self.current_node_users} utilisateurs connectes")
                    self.broadcast(f"MSG:{pseudo}#{uid} a rejoint le noeud.", pseudo)
                elif msg_type == "LEAVE":
                    with self.lock:
                        if addr in self.users:
                            del self.users[addr]
                            self.current_node_users -= 1
                            print(f"-> {self.users[addr].pseudo} a quitte le noeud.")
                            print(f"Il y a {self.current_node_users} utilisateurs connectes")
                            left_pseudo = self.users.get(addr, User(None, None, "Inconnu", "??")).pseudo
                            self.broadcast(f"MSG:{left_pseudo} a quitte le noeud.", None)
                    conn.close()
                elif msg_type == "MSG":
                    msg_content = ':'.join(parts[3:])
                    print(f"[{pseudo}#{uid}]: {msg_content}")
                    self.broadcast(f"MSG:{pseudo}#{uid}: {msg_content}", pseudo)
        except Exception as e:
            pass
        finally:
            try:
                conn.close()
            except:
                pass

    def broadcast(self, message, sender_pseudo):
        with self.lock:
            for addr, user in list(self.users.items()):
                try:
                    user.conn.sendall((message + "\n").encode())
                except:
                    pass

    def connect_to_node(self, host, port):
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.connect((host, port))
            join_msg = f"{self.local_pseudo}:{self.local_uid}:JOIN"
            sock.sendall(join_msg.encode())
            
            print("Vous etes tout seul, merci d'attendre...")
            
            while self.running:
                try:
                    data = sock.recv(1024).decode()
                    if not data:
                        break
                    
                    if "a rejoint" in data or "a quitte" in data:
                        print(data)
                    else:
                        print(data)
                except:
                    break
            
            leave_msg = f"{self.local_pseudo}:{self.local_uid}:LEAVE"
            sock.sendall(leave_msg.encode())
            sock.close()
            print("Deconnecte du noeud.")
        except Exception as e:
            print(f"Erreur de connexion : {e}")

    def start_irc(self):
        print("--- Mode IRC ---")
        irc_host = input("Serveur IRC (ex: irc.libera.chat) : ").strip()
        irc_port = 6667
        try:
            irc_port = int(input("Port (defaut 6667) : ").strip() or "6667")
        except:
            pass
            
        irc_nick = self.local_pseudo
        print(f"Connexion a {irc_host}:{irc_port} en tant que {irc_nick}...")
        
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.connect((irc_host, irc_port))
            
            sock.sendall(f"NICK {irc_nick}\n".encode())
            sock.sendall(f"USER pytopy 8 * :pytopy client\n".encode())
            channel = input(f"Canal a rejoindre (ex: {irc_nick}): ").strip()
            if not channel.startswith("#"):
                channel = "#" + channel
            sock.sendall(f"JOIN {channel}\n".encode())
            
            print(f"Connecte sur {channel}. Tapez 'QUIT' pour sortir.")

            def recv_loop():
                while self.running:
                    try:
                        data = sock.recv(4096).decode()
                        if not data:
                            break
                        print(data.strip().replace("\r", ""))
                    except:
                        break

            t_recv = threading.Thread(target=recv_loop, daemon=True)
            t_recv.start()

            while self.running:
                cmd = input("> ").strip()
                if cmd.upper() == "QUIT":
                    break
                if cmd:
                    sock.sendall(f"PRIVMSG {channel} :{cmd}\n".encode())
            
            sock.sendall(b"QUIT :pytopy bye\n")
            sock.close()
        except Exception as e:
            print(f"Erreur IRC : {e}")

    def run(self):
        self.choose_identity()
        self.server_thread = threading.Thread(target=self.start_server, daemon=True)
        self.server_thread.start()
        
        while self.running:
            print("\n--- Menu pytopy ---")
            print("1. Creer/Rejoindre un noeud P2P")
            print("2. Mode IRC")
            print("3. Quitter")
            
            choice = input("Choix : ").strip()
            
            if choice == "1":
                mode = input("1 pour Creer, 2 pour Rejoindre : ").strip()
                if mode == "1":
                    print("Le noeud est cree. Attendez les connexions.")
                    time.sleep(1)
                elif mode == "2":
                    host = input("Adresse du noeud (IP ou localhost) : ").strip()
                    port = 9999
                    try:
                        port = int(input("Port (defaut 9999) : ").strip() or "9999")
                    except:
                        pass
                    self.connect_to_node(host, port)
                else:
                    print("Choix invalide.")
            elif choice == "2":
                self.start_irc()
            elif choice == "3":
                self.running = False
                os.system('clear')
                sys.exit(0)
            else:
                print("Choix invalide.")

if __name__ == "__main__":
    node = P2PNode()
    try:
        node.run()
    except KeyboardInterrupt:
        node.running = False
        print("\nArret de pytopy.")
